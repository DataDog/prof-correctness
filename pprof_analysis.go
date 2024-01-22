package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	PprofRegex   string         `json:"pprof-regex"`
	StackContent []StackContent `json:"stack-content"`
}

type StackTestData struct {
	TestName string        `json:"test_name"`
	Stacks   []TypedStacks `json:"stacks"`
}

func (stack *StackContent) UnmarshalJSON(data []byte) error {
	type stackcontent StackContent
	stackContent := &stackcontent{
		Value:   -1, // default value
		Percent: -1, // default value
	}

	err := json.Unmarshal(data, stackContent)
	if err != nil {
		return err
	}

	*stack = StackContent(*stackContent)
	return nil
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

func captureProfData(t *testing.T, prof *profile.Profile, path string, testName string) {
	// labels to ignore
	keysToIgnore := []string{"thread id", "thread native id"}

	var capturedData StackTestData
	capturedData.TestName = testName

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
				matched, err := regexp.MatchString(v, vals[i])
				if err != nil {
					t.Fatalf("Error matching regexp %s: %v", v, err)
				}
				if !matched {
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

func assertStack(t *testing.T, prof []StackSample, regexpStack string, value float64, pct int64, epsilonPct int64, labels []Labels) {
	r, err := regexp.Compile(regexpStack)
	if err != nil {
		t.Fatalf("Error compiling regex: %v, %s", err, regexpStack)
	}
	var total int64 = 0
	var matching int64 = 0
	for _, ss := range prof {
		total += ss.Val
		if r.MatchString(ss.Stack) {
			if labels == nil || checkLabels(t, ss.Labels, labels) {
				matching += ss.Val
			}
		}
	}

	var actualPct int64 = 0
	if total != 0 {
		actualPct = matching * 100 / total
	}

	if value >= 0 {
		errorPct := math.Abs(float64(matching)-value) / math.Max(value, math.SmallestNonzeroFloat64) * 100.0
		// t.Logf("Stack '%s' should be %d%% +/- %d%% of the profile and is %d%%\n", stack, pct, epsilonPct, actualPct)
		if errorPct > float64(epsilonPct) {
			t.Errorf("\033[31mAssertion failed: stack '%s' should have been %.1f +/- %d%% of the profile but was %d with %.1f%% error\033[0m", regexpStack, value, epsilonPct, matching, errorPct)
		} else {
			t.Logf("\033[32mAssertion succeeded: stack '%s' is %.1f +/- %d%% of the profile (was %d with %.1f%% error)\033[0m", regexpStack, value, epsilonPct, matching, errorPct)
		}
	}

	if pct >= 0 {
		diff := absDiff(pct, actualPct)
		// t.Logf("Stack '%s' should be %d%% +/- %d%% of the profile and is %d%%\n", stack, pct, epsilonPct, actualPct)
		if diff > epsilonPct {
			t.Errorf("\033[31mAssertion failed: stack '%s' should have been %d%% +/- %d%% of the profile but was %d%% with %d%% error\033[0m", regexpStack, pct, epsilonPct, actualPct, diff)
		} else {
			t.Logf("\033[32mAssertion succeeded: stack '%s' is %d%% +/- %d%% of the profile (was %d%% with %d%% error)\033[0m", regexpStack, pct, epsilonPct, actualPct, diff)
		}
	}
}

func analyzeProfData(t *testing.T, prof []StackSample, typedStacks TypedStacks, durationSecs float64) {
	for _, stack := range typedStacks.StackContent {
		regexpStack := stack.RegularExpression
		// Do not scale values for profiles with a duration of 0 (eg. Node.js heap profiles)
		value := float64(stack.Value)
		if durationSecs > 0 {
			value = value * durationSecs // value for total duration
		}
		percent := stack.Percent // percentage within the profile
		errorMargin := stack.ErrorMargin
		assertStack(t, prof, regexpStack, value, percent, errorMargin, stack.Labels)
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

func readJSONFile(filePath string) (StackTestData, error) {
	var data StackTestData
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return data, err
	}
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return data, err
	}
	if !json.Valid(byteValue) {
		return data, fmt.Errorf("Invalid json data %s", filePath)
	}
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return data, err
	}
	return data, nil
}

func getMatchingFiles(folder string, filenameRegex *regexp.Regexp) ([]string, error) {
	var matchingFiles []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filenameRegex.MatchString(info.Name()) {
			matchingFiles = append(matchingFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matchingFiles, nil
}

func readPprofFile(pprof_file string) (*profile.Profile, error) {
	// Open the file
	file, err := os.Open(pprof_file)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if ok, _ := lz4.ValidFrameHeader(content); ok {
		in := bytes.NewReader(content)
		zr := lz4.NewReader(in)
		var out bytes.Buffer
		// is lz4 compressed? lets decompress that
		_, err := io.Copy(&out, zr)
		if err != nil {
			return nil, err
		}
		content = out.Bytes()
		zr.Reset(nil)
	}
	prof, err := profile.ParseData(content)
	if err != nil {
		return nil, err
	}
	return prof, nil
}

func analyzePprofFile(t *testing.T, pprof_file string, typedStacks TypedStacks, testName string, captureData bool) {
	prof, err := readPprofFile(pprof_file)
	if err != nil {
		t.Fatalf("Error reading file %s", pprof_file)
	}
	t.Logf("Analyzing results in %s for profile type %s", pprof_file, typedStacks.ProfileType)

	profileDuration := float64(prof.DurationNanos) / 1000000000.0
	t.Logf("Found a profile duration of %.1f seconds (in %s)", profileDuration, filepath.Base(pprof_file))

	// Store current data in a json file to help users create their tests
	if captureData {
		captureProfData(t, prof, pprof_file, testName)
	}

	typedProf := getProfileType(t, prof, typedStacks.ProfileType)
	analyzeProfData(t, typedProf, typedStacks, profileDuration)
}

func AnalyzeResults(t *testing.T, jsonFilePath string, pprof_folder string) {
	stackTestData, err := readJSONFile(jsonFilePath)
	if err != nil {
		t.Fatalf("Error opening file %s", jsonFilePath)
	}

	// python files are in the form "profile.<pid>.number"
	// Other profilers (using pprof) include pprof in the name
	// Filter out files that ends with '.json' to avoid considering files dumped by captureProfData as profiles
	default_pprof_regexp := regexp.MustCompile("(^profile.*|.*pprof.*)([^n]|[^o]n|[^s]on|[^j]son|[^.]json)$")
	processedProfilesMap := make(map[string]bool)

	for _, typedStacks := range stackTestData.Stacks {
		// use typedStack.PprofRegex if defined, otherwise use default_pprof_regexp
		pprof_regexp := default_pprof_regexp
		if typedStacks.PprofRegex != "" {
			pprof_regexp = regexp.MustCompile(typedStacks.PprofRegex)
		}
		matchingFiles, err := getMatchingFiles(pprof_folder, pprof_regexp)
		if err != nil {
			t.Fatalf("Error getting matching files: %v", err)
		}
		if len(matchingFiles) == 0 {
			t.Errorf("No matching files found for %s in %s", pprof_regexp, pprof_folder)
		} else {
			for _, file := range matchingFiles {
				_, fileAlreadyProcessed := processedProfilesMap[file]
				if !fileAlreadyProcessed {
					processedProfilesMap[file] = true
				}
				analyzePprofFile(t, file, typedStacks, stackTestData.TestName, !fileAlreadyProcessed)
			}
		}
	}
}
