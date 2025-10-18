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
	"strconv"
	"strings"
	"testing"

	"github.com/google/pprof/profile"
	"github.com/pierrec/lz4/v4"
	"github.com/klauspost/compress/zstd"
)

var (
	_ json.Unmarshaler = (*Optional[int64])(nil)
	_ json.Marshaler   = (*Optional[int64])(nil)
)

type Optional[T any] struct {
	value *T
}

func NewOptionalFrom[T any](v T) (o Optional[T]) {
	o.value = &v
	return
}

func (o *Optional[T]) UnmarshalJSON(bytes []byte) error {
	o.value = new(T)
	return json.Unmarshal(bytes, o.value)
}

func (o *Optional[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.value)
}

func (o *Optional[T]) Value() (out T, ok bool) {
	if o.value == nil {
		return
	}
	return *o.value, true
}

func MapOptional[I any, O any](option Optional[I], mapper func(v I) O) (mappedOption Optional[O]) {
	if v, ok := option.Value(); ok {
		mappedV := mapper(v)
		mappedOption.value = &mappedV
	}
	return
}

type StackSample struct {
	Stack  string // folded-style: func1;func2;func3
	Val    int64
	Labels map[string][]string
}

// Reference data from the json files
type Labels struct {
	Key         string   `json:"key"`
	Values      []string `json:"values"`       // fixed value
	ValuesRegex string   `json:"values_regex"` // regex for values
}

type StackContent struct {
	RegularExpression string `json:"regular_expression"`
	// NOTE: When the corresponding profile has a duration > 0, this value represents a rate (x/sec).
	//       If the corresponding profile is a snapshot (i.e. duration == 0), then this value represents
	//       an absolute/raw/scalar value independent of time.
	Value       Optional[int64] `json:"value"`
	Percent     Optional[int64] `json:"percent"`
	ErrorMargin Optional[int64] `json:"error_margin,omitempty"`
	Labels      []Labels        `json:"labels"`
}

type TypedStacks struct {
	ProfileType  string         `json:"profile-type"`
	PprofRegex   string         `json:"pprof-regex"`
	StackContent []StackContent `json:"stack-content"`
	ErrorMargin  int64          `json:"error-margin,omitempty"`
	// NOTE: When the corresponding profile has a duration > 0, this value represents a rate (x/sec).
	//       If the corresponding profile is a snapshot (i.e. duration == 0), then this value represents
	//       an absolute/raw/scalar value independent of time.
	ValueMatchingSum Optional[int64] `json:"value-matching-sum,omitempty"`
}

type StackTestData struct {
	TestName        string        `json:"test_name"`
	ScaleByDuration bool          `json:"scale_by_duration"`
	PprofRegex      string        `json:"pprof-regex"`
	Stacks          []TypedStacks `json:"stacks"`
}

// Custom unmarshaller for Labels to ensure exactly one of Values and ValueRegex is defined
func (l *Labels) UnmarshalJSON(data []byte) error {
	type labels Labels
	var tmp labels
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if (tmp.Values != nil) == (tmp.ValuesRegex != "") {
		return fmt.Errorf("Exactly one of values and value_regex must be defined")
	}

	sort.Strings(tmp.Values)

	*l = Labels(tmp)
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

func relDiff(actual, reference float64) float64 {
	return math.Abs((actual - reference) / math.Max(reference, math.SmallestNonzeroFloat64) * 100.0)
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

func captureProfData(t *testing.T, prof *profile.Profile, path string, testName string, profileDuration float64) {
	// labels to ignore
	keysToIgnore := []string{"thread native id"}

	var capturedData StackTestData
	capturedData.TestName = testName

	for _, sampleType := range prof.SampleType {
		var typedStack TypedStacks
		typedStack.ProfileType = sampleType.Type
		typedStack.ErrorMargin = 1

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

			if profileDuration > 0 {
				// NOTE: When profile duration is bigger than 0, all values represent rates.
				ss.Val = int64(float64(ss.Val) / profileDuration)
			}

			stackContent := StackContent{
				Value: NewOptionalFrom(ss.Val),
				// protect any charact
				RegularExpression: "^" + regexp.QuoteMeta(ss.Stack) + "$",
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
				if val, ok := typedStack.StackContent[idx].Value.Value(); ok {
					pct := (val * 100) / int64(totalVal)
					typedStack.StackContent[idx].Percent = NewOptionalFrom(pct)
					if pct < typedStack.ErrorMargin {
						idxToRemove = append(idxToRemove, idx)
					}
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
	if err := profile.Aggregate(true, true, false, false, false, false); err != nil {
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
		for k, v := range sample.NumLabel {
			for _, i := range v {
				labels[k] = append(labels[k], strconv.FormatInt(i, 10))
			}
			sort.Strings(labels[k])
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

func checkLabels(t *testing.T, labels map[string][]string, expectedLabels []Labels) bool {
	for _, expectedLabel := range expectedLabels {
		if values, ok := labels[expectedLabel.Key]; ok {
			if expectedLabel.Values != nil {
				// Right now all values should be present.
				// t.Log("Checking: vals ", values, "vs ", expectedLabel.Values, "key", expectedLabel.Key)
				if len(values) != len(expectedLabel.Values) {
					// t.Log("NO")
					return false
				}
				// Sample values and exepected values are sorted when read from profile/json file
				for i, v := range expectedLabel.Values {
					if values[i] != v {
						return false
					}
				}
			} else {
				// Sample values and expected values are sorted when read from profile/json file
				for _, v := range values {
					matched, err := regexp.MatchString(expectedLabel.ValuesRegex, v)
					if err != nil {
						t.Fatalf("Error matching regexp %s: %v", v, err)
					}
					if !matched {
						return false
					}
				}
			}
		} else {
			return false
		}
	}
	return true
}

func assertStack(t *testing.T, prof []StackSample, regexpStack string, valueOpt Optional[float64], pctOpt Optional[int64], epsilonPct int64, labels []Labels) (matching int64) {
	r, err := regexp.Compile(regexpStack)
	if err != nil {
		t.Fatalf("Error compiling regex: %v, %s", err, regexpStack)
	}
	var total int64 = 0
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

	if value, ok := valueOpt.Value(); ok {
		errorPct := relDiff(float64(matching), value)
		// t.Logf("Stack '%s' should be %d%% +/- %d%% of the profile and is %d%%\n", stack, pct, epsilonPct, actualPct)
		if errorPct > float64(epsilonPct) {
			t.Errorf("\033[31mAssertion failed: stack '%s' (labels=%v) should have been %.1f +/- %d%% of the profile but was %d with %.1f%% error\033[0m", regexpStack, labels, value, epsilonPct, matching, errorPct)
		} else {
			t.Logf("\033[32mAssertion succeeded: stack '%s' (labels=%v) is %.1f +/- %d%% of the profile (was %d with %.1f%% error)\033[0m", regexpStack, labels, value, epsilonPct, matching, errorPct)
		}
	}

	if pct, ok := pctOpt.Value(); ok {
		diff := absDiff(pct, actualPct)
		// t.Logf("Stack '%s' should be %d%% +/- %d%% of the profile and is %d%%\n", stack, pct, epsilonPct, actualPct)
		if diff > epsilonPct {
			t.Errorf("\033[31mAssertion failed: stack '%s' (labels=%v) should have been %d%% +/- %d%% of the profile but was %d%% with %d%% error\033[0m", regexpStack, labels, pct, epsilonPct, actualPct, diff)
		} else {
			t.Logf("\033[32mAssertion succeeded: stack '%s' (labels=%v) is %d%% +/- %d%% of the profile (was %d%% with %d%% error)\033[0m", regexpStack, labels, pct, epsilonPct, actualPct, diff)
		}
	}
	return
}

func analyzeProfData(t *testing.T, prof []StackSample, typedStacks TypedStacks, durationSecs float64) {
	var matchingSum int64 = 0
	for _, stack := range typedStacks.StackContent {
		regexpStack := stack.RegularExpression
		// Do not scale values for profiles with a duration of 0 (eg. Node.js heap profiles)
		valueOpt := MapOptional(stack.Value, func(v int64) float64 { return float64(v) })
		if durationSecs > 0 {
			// NOTE: When profile duration is bigger than 0, all values represent rates.
			valueOpt = MapOptional(valueOpt, func(v float64) float64 { return v * durationSecs }) // value for total duration
		}
		percent := stack.Percent // percentage within the profile

		errorMargin := typedStacks.ErrorMargin
		if stackErrorMargin, ok := stack.ErrorMargin.Value(); ok {
			errorMargin = stackErrorMargin
		}

		matchingSum += assertStack(t, prof, regexpStack, valueOpt, percent, errorMargin, stack.Labels)
		// todo
		// - add an assertion on counts (example:number of allocations)
	}

	if expectedSum, ok := typedStacks.ValueMatchingSum.Value(); ok {
		value := float64(expectedSum)
		if durationSecs > 0 {
			// NOTE: When profile duration is bigger than 0, all values represent rates.
			value = value * durationSecs
		}
		errorPct := relDiff(float64(matchingSum), value)
		if errorPct > float64(typedStacks.ErrorMargin) {
			t.Errorf("\033[31mAssertion failed: profile '%s' should have total matching sum of %1.f +/- %d%% but was %d with %.1f%% error\033[0m", typedStacks.ProfileType, value, typedStacks.ErrorMargin, matchingSum, errorPct)
		} else {
			t.Logf("\033[32mAssertion succeeded: profile '%s' has total matching sum of %1.f +/- %d%% (was %d with %.1f%% error)\033[0m", typedStacks.ProfileType, value, typedStacks.ErrorMargin, matchingSum, errorPct)
		}
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

	// handle zstd compressed profiles (magic bytes 0x28B52FFD / 0xFD2FB528)
	if len(content) >= 4 {
		magic := uint32(content[0]) | uint32(content[1])<<8 | uint32(content[2])<<16 | uint32(content[3])<<24
		if magic == 0xFD2FB528 || magic == 0x28B52FFD { // little or big endian check
			dec, err := zstd.NewReader(nil)
			if err != nil {
				return nil, err
			}
			defer dec.Close()
			decompressed, err := dec.DecodeAll(content, nil)
			if err != nil {
				return nil, err
			}
			content = decompressed
		}
	}
	prof, err := profile.ParseData(content)
	if err != nil {
		return nil, err
	}
	return prof, nil
}

func analyzePprofFile(t *testing.T, pprof_file string, typedStacks TypedStacks, testName string, captureData bool, scaleByDuration bool) {
	prof, err := readPprofFile(pprof_file)
	if err != nil {
		t.Fatalf("Error reading file %s", pprof_file)
	}
	t.Logf("Analyzing results in %s for profile type %s", pprof_file, typedStacks.ProfileType)

	profileDuration := float64(prof.DurationNanos) / 1000000000.0
	t.Logf("Found a profile duration of %.1f seconds (in %s)", profileDuration, filepath.Base(pprof_file))

	// Store current data in a json file to help users create their tests
	if captureData {
		captureProfData(t, prof, pprof_file, testName, profileDuration)
	}
	if !scaleByDuration {
		// ignore duration, values can be considered absolute
		profileDuration = 0
	}
	typedProf := getProfileType(t, prof, typedStacks.ProfileType)
	analyzeProfData(t, typedProf, typedStacks, profileDuration)
}

func AnalyzeResults(t *testing.T, jsonFilePath string, pprof_folder string) {
	stackTestData, err := readJSONFile(jsonFilePath)
	if err != nil {
		t.Fatalf("Error opening file %s: %v", jsonFilePath, err)
	}

	var default_pprof_regexp *regexp.Regexp
	if stackTestData.PprofRegex != "" {
		default_pprof_regexp = regexp.MustCompile(stackTestData.PprofRegex)
	} else {
		// python files are in the form "profile.<pid>.number"
		// Other profilers (using pprof) include pprof in the name
		// Filter out files that ends with '.json' to avoid considering files dumped by captureProfData as profiles
		// Golang regexes do not have negative lookahed, so we need to use `([^n]|[^o]n|[^s]on|[^j]son|[^.]json)$` instead of `(?![.]json)$
		default_pprof_regexp = regexp.MustCompile("^(profile|.*pprof)($|.*([^n]|[^o]n|[^s]on|[^j]son|[^.]json)$)")
	}
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
				analyzePprofFile(t, file, typedStacks, stackTestData.TestName, !fileAlreadyProcessed, stackTestData.ScaleByDuration)
			}
		}
	}
}
