package main

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/pprof/profile"
	"github.com/pierrec/lz4/v4"
)

type StackSample struct {
	Stack  string // folded-style: func1;func2;func3
	Val    int64
	Labels map[string][]string
}

// Reference data from the json files
type Labels struct {
	Key    string   `json:"key"`
	Values []string `json:"values"` // fixed value
}

type StackContent struct {
	RegularExpression string   `json:"regular_expression"`
	Value             int64    `json:"value"`
	Percent           int64    `json:"percent"`
	ErrorMargin       int64    `json:"error_margin"`
	Labels            []Labels `json:"labels"`
}

type TypedStacks struct {
	ProfileType  string         `json:"profile-type"`
	StackContent []StackContent `json:"stack-content"`
}

type StackTestData struct {
	TestName string        `json:"test_name"`
	Stacks   []TypedStacks `json:"stacks"`
}

func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func absDiff(x, y int64) int64 {
	if x < y {
		return y - x
	}
	return x - y
}

func contains(s []int, v int) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func contains_str(s []string, v string) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func captureProfData(t *testing.T, prof *profile.Profile, path string, stackTestData StackTestData) {
	// labels to ignore
	keysToIgnore := []string{"thread id", "thread native id"}

	var capturedData StackTestData
	capturedData.TestName = stackTestData.TestName

	for _, sampleType := range prof.SampleType {
		var typedStack TypedStacks
		typedStack.ProfileType = sampleType.Type

		typedProf := getProfileType(t, prof, sampleType.Type)
		// drop the content to a file to allow a comparison
		var totalVal int = 0
		for _, ss := range typedProf {
			var labels []Labels

			for key, value := range ss.Labels {
				if contains_str(keysToIgnore, key) {
					continue
				}
				labels = append(labels, Labels{
					Key:    key,
					Values: value,
				})
			}

			stackContent := StackContent{
				ErrorMargin: 3, // TODO: is this a good default ?
				Value:       ss.Val,
				// protect any charact
				RegularExpression: regexp.QuoteMeta(ss.Stack),
				Labels:            labels,
			}
			typedStack.StackContent = append(typedStack.StackContent, stackContent)
			totalVal += int(ss.Val)
		}
		// filter out and add significance (%)
		var newStackContent []StackContent
		if totalVal != 0 {
			var idxToRemove []int
			for idx := range typedStack.StackContent {
				typedStack.StackContent[idx].Percent = (typedStack.StackContent[idx].Value * 100) / int64(totalVal)
				if typedStack.StackContent[idx].Percent < typedStack.StackContent[idx].ErrorMargin {
					idxToRemove = append(idxToRemove, idx)
				}
			}
			// rebuild a new table without the elements that have a low percentage
			for idx, content := range typedStack.StackContent {
				if !contains(idxToRemove, idx) {
					newStackContent = append(newStackContent, content)
				}
			}
		}
		typedStack.StackContent = newStackContent

		capturedData.Stacks = append(capturedData.Stacks, typedStack)
	}

	jsonPath := filepath.Join(filepath.Dir(path), fileNameWithoutExt(filepath.Base(path))) + ".json"

	err := writeToJSONFile(capturedData, jsonPath)
	if err != nil {
		t.Fatalf("Failed to write : %v", err)
	} else {
		t.Log("Results stored in ", jsonPath)
	}
}

func getProfileType(t *testing.T, profile *profile.Profile, type_ string) []StackSample {
	typeIdx := -1
	for i, sampleType := range profile.SampleType {
		if sampleType.Type == type_ {
			typeIdx = i
		}
	}
	if typeIdx == -1 {
		t.Fatalf("Couldn't find sample type %s", type_)
	}
	// t.Logf("Found '%s' smaple type at idx %d\n", type_, typeIdx)

	// if err := profile.Aggregate(true, true, false, p.LineNumbers, false); err != nil {
	if err := profile.Aggregate(true, true, false, false, false); err != nil {
		t.Fatalf("Error aggregating profile samples: %v", err)
	}
	profile = profile.Compact()
	sort.Slice(profile.Sample, func(i, j int) bool {
		return profile.Sample[i].Value[0] > profile.Sample[j].Value[0]
	})

	var out []StackSample
	for _, sample := range profile.Sample {
		var frames []string
		for i := range sample.Location {
			loc := sample.Location[len(sample.Location)-i-1]
			for j := range loc.Line {
				line := loc.Line[len(loc.Line)-j-1]
				name := line.Function.Name
				// if p.LineNumbers {
				//     name = name + ":" + strconv.FormatInt(line.Line, 10)
				// }
				frames = append(frames, name)
			}
		}
		labels := make(map[string][]string)
		for k, v := range sample.Label {
			// ease the comparison by sorting string values
			sort.Strings(v)
			labels[k] = v
			// t.Log("Sorted labels :", v)
		}
		ss := StackSample{
			Stack:  strings.Join(frames, ";"),
			Val:    sample.Value[typeIdx],
			Labels: labels,
		}
		out = append(out, ss)
	}
	return out
}

func checkLabels(t *testing.T, labels map[string][]string, expected []Labels) bool {
	for _, e := range expected {
		if vals, ok := labels[e.Key]; ok {
			// Right now all values should be present.
			// t.Log("Checking: vals ", vals, "vs ", e.Values, "key", e.Key)
			if len(vals) != len(e.Values) {
				// t.Log("NO")
				return false
			}
			// Sample values for labels are sorted when read from stacks
			sort.Strings(e.Values)
			for i, v := range e.Values {
				if vals[i] != v {
					// t.Log("NO")
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func assertStackPercent(t *testing.T, prof []StackSample, regexpStack string, pct int64, epsilonPct int64, labels []Labels) {
	r, err := regexp.Compile(regexpStack)
	if err != nil {
		t.Fatalf("Error compiling regex: %v, %s", err, regexpStack)
	}
	var total int64 = 0
	var matching int64 = 0
	var found bool = false
	for _, ss := range prof {
		total += ss.Val
		if r.MatchString(ss.Stack) {
			if labels == nil || checkLabels(t, ss.Labels, labels) {
				matching += ss.Val
				found = true
			}
		}
	}

	if !found {
		t.Errorf("Assertion failed: stack '%s' not found", regexpStack)
		return
	}

	var actualPct int64 = 0
	if total != 0 {
		actualPct = matching * 100 / total
	}

	diff := absDiff(pct, actualPct)
	// t.Logf("Stack '%s' should be %d%% +/- %d%% of the profile and is %d%%\n", stack, pct, epsilonPct, actualPct)
	if diff > epsilonPct {
		t.Errorf("Assertion failed: stack '%s' should have been %d%% +/- %d%% of the profile but was %d%% with %d%% error", regexpStack, pct, epsilonPct, actualPct, diff)
	} else {
		t.Logf("Assertion succeeded: stack '%s' is %d%% +/- %d%% of the profile (was %d%% with %d%% error)", regexpStack, pct, epsilonPct, actualPct, diff)
	}
}

func assertStackValue(t *testing.T, prof []StackSample, regexpStack string, value float64, epsilonPct int64) {
	r, err := regexp.Compile(regexpStack)
	if err != nil {
		t.Fatalf("Error compiling regex: %v, %s", err, regexpStack)
	}
	var matching int64 = 0
	for _, ss := range prof {
		if r.MatchString(ss.Stack) {
			matching += ss.Val
		}
	}

	errorPct := math.Abs(float64(matching)-value) / value * 100.0
	// t.Logf("Stack '%s' should be %d%% +/- %d%% of the profile and is %d%%\n", stack, pct, epsilonPct, actualPct)
	if errorPct > float64(epsilonPct) {
		t.Errorf("Assertion failed: stack '%s' should have been %.1f +/- %d%% of the profile but was %d with %.1f%% error", regexpStack, value, epsilonPct, matching, errorPct)
	} else {
		t.Logf("Assertion succeeded: stack '%s' is %.1f +/- %d%% of the profile (was %d with %.1f%% error)", regexpStack, value, epsilonPct, matching, errorPct)
	}
}

func analyzeProfData(t *testing.T, prof []StackSample, typedStacks TypedStacks, durationSecs float64) {
	for _, stack := range typedStacks.StackContent {
		regexpStack := stack.RegularExpression
		value := float64(stack.Value) * durationSecs // value for total duration
		percent := stack.Percent                     // percentage within the profile
		errorMargin := stack.ErrorMargin
		if percent != 0 {
			assertStackPercent(t, prof, regexpStack, percent, errorMargin, stack.Labels)
		}

		if value != 0 {
			assertStackValue(t, prof, regexpStack, value, errorMargin)
		}
		// todo
		// - add an assertion on counts (example:number of allocations)
		// - add an assertion on the total amount captured
	}
}

func writeToJSONFile(data StackTestData, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, 0644)
}

func AnalyzeResults(t *testing.T, jsonFilePath string, pprof_folder string) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		t.Fatalf("Error opening file %s", jsonFilePath)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		t.Fatalf("Unable to read json data %s", jsonFilePath)
	}

	if !json.Valid(byteValue) {
		t.Fatalf("Invalid json data %s", jsonFilePath)
	}

	var stackTestData StackTestData
	if err := json.Unmarshal(byteValue, &stackTestData); err != nil {
		t.Fatalf("Unable to Unmarshal json data %s", jsonFilePath)
	}

	found_pprof := false
	// retrieve all stack data
	stacks := stackTestData.Stacks
	// python files are in the form "profile.<pid>.number"
	// Other profilers (using pprof) include pprof in the name
	pprof_regexp := regexp.MustCompile("(^profile.*|.*pprof.*)")

	zr := lz4.NewReader(nil)

	// Iterate over all files in the pprof folder
	filepath.Walk(pprof_folder, func(path string, info os.FileInfo, err error) error {
		// anon function that opens all the prof data and checks that it has the correct stacks
		if err != nil {
			t.Fatalf("Error walking pprof folder: %v", err)
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Only consider pprof files
		if !pprof_regexp.MatchString(filepath.Base(path)) {
			return nil
		}
		// Open the file
		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("Error opening file %s", path)
		}
		defer file.Close()
		// Read the file content
		t.Logf("Analyzing results in %s", path)
		content, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Error reading file %s", path)
		}
		if ok, _ := lz4.ValidFrameHeader(content); ok {
			in := bytes.NewReader(content)
			zr.Reset(in)
			var out bytes.Buffer
			// is lz4 compressed? lets decompress that
			_, err := io.Copy(&out, zr)
			if err != nil {
				t.Fatalf("Failed to decompress lz4 pprof: %v", err)
			}
			content = out.Bytes()
			zr.Reset(nil)
		}
		prof, err := profile.ParseData(content)
		if err != nil {
			t.Fatalf("Failed to parse profile: %v", err)
		}
		found_pprof = true
		profileDuration := float64(prof.DurationNanos) / 1000000000.0
		t.Logf("Found a profile duration of %.1f seconds (in %s)", profileDuration, filepath.Base(path))

		// Store current data in a json file to help users create their tests
		captureProfData(t, prof, path, stackTestData)

		// Loop on all profile types
		for _, typedStacks := range stacks {
			typedProf := getProfileType(t, prof, typedStacks.ProfileType)

			analyzeProfData(t, typedProf, typedStacks, profileDuration)
		}
		return nil
	})
	if !found_pprof {
		t.Fatalf("No pprof file found. Check what profiler emitted")
	}
}
