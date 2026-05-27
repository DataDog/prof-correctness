// Package analysis parses pprof profiles and asserts their contents against an
// expected_profile.json description. It is decoupled from the testing package
// (via the Reporter interface) so it can be reused by non-Go-test runners,
// such as a Windows scenario harness in another repository.
package analysis

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

	"github.com/google/pprof/profile"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/xeipuuv/gojsonschema"
)

var (
	_ json.Unmarshaler = (*Optional[int64])(nil)
	_ json.Marshaler   = (*Optional[int64])(nil)
)

// JSON Schema for validating expected profile JSON files.
// Basic structure validation, complex rules validated in Go code.
var expectedProfileSchema = `{
  "$schema": "https://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["stacks"],
  "properties": {
    "test_name": { "type": "string" },
    "note": { "type": "string" },
    "scale_by_duration": { "type": "boolean" },
    "pprof-regex": { "type": "string" },
    "allow_first_profile_failure": { "type": "boolean" },
    "stacks": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["profile-type", "stack-content"],
        "properties": {
          "profile-type": { "type": "string", "minLength": 1 },
          "pprof-regex": { "type": "string" },
          "stack-content": {
            "type": "array",
            "minItems": 1,
            "items": {
              "type": "object",
              "required": ["regular_expression"],
              "properties": {
                "regular_expression": { "type": "string", "minLength": 1 },
                "value": { "type": "integer" },
                "percent": { "type": "integer" },
                "error_margin": { "type": "integer" },
                "labels": { "type": "array" }
              }
            }
          },
          "error-margin": { "type": "integer" },
          "value-matching-sum": { "type": "integer" }
        }
      }
    }
  }
}`

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
	TestName                 string        `json:"test_name"`
	Note                     string        `json:"note,omitempty"`
	ScaleByDuration          bool          `json:"scale_by_duration"`
	PprofRegex               string        `json:"pprof-regex"`
	AllowFirstProfileFailure bool          `json:"allow_first_profile_failure,omitempty"`
	Stacks                   []TypedStacks `json:"stacks"`
}

// Validate rules that JSON Schema can't express
func (s *StackTestData) Validate() error {
	// Stacks must be non-empty unless note is present
	if len(s.Stacks) == 0 && s.Note == "" {
		return fmt.Errorf("'stacks' must have at least one entry (or provide a 'note' explaining why it's empty)")
	}

	// If no value-matching-sum, require value or percent in stack-content
	for i, stack := range s.Stacks {
		if _, hasValueMatchingSum := stack.ValueMatchingSum.Value(); hasValueMatchingSum {
			continue
		}
		for j, content := range stack.StackContent {
			_, hasValue := content.Value.Value()
			_, hasPercent := content.Percent.Value()
			if !hasValue && !hasPercent {
				return fmt.Errorf("stacks[%d].stack-content[%d]: must have 'value' or 'percent' (or parent must have 'value-matching-sum')", i, j)
			}
		}
	}

	return nil
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

func containsStr(s []string, v string) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func captureProfData(r Reporter, prof *profile.Profile, path string, testName string, profileDuration float64) {
	// labels to ignore
	keysToIgnore := []string{"thread native id"}

	var capturedData StackTestData
	capturedData.TestName = testName

	for _, sampleType := range prof.SampleType {
		var typedStack TypedStacks
		typedStack.ProfileType = sampleType.Type
		typedStack.ErrorMargin = 1

		typedProf := getProfileType(r, prof, sampleType.Type)
		// drop the content to a file to allow a comparison
		var totalVal int = 0
		for _, ss := range typedProf {
			var labels []Labels

			for key, value := range ss.Labels {
				if containsStr(keysToIgnore, key) {
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
		// Annotate each entry with its percentage of the total. No filtering —
		// users (and LLMs) curating an expected_profile.json from this output
		// want to see every stack, including the long tail, so they can decide
		// what's meaningful instead of having that decided for them.
		if totalVal != 0 {
			for idx := range typedStack.StackContent {
				if val, ok := typedStack.StackContent[idx].Value.Value(); ok {
					pct := (val * 100) / int64(totalVal)
					typedStack.StackContent[idx].Percent = NewOptionalFrom(pct)
				}
			}
		}

		capturedData.Stacks = append(capturedData.Stacks, typedStack)
	}

	jsonPath := filepath.Join(filepath.Dir(path), fileNameWithoutExt(filepath.Base(path))) + ".json"

	err := writeToJSONFile(capturedData, jsonPath)
	if err != nil {
		r.Fatalf("Failed to write : %v", err)
	} else {
		r.Logf("Results stored in %s", jsonPath)
	}
}

func getProfileType(r Reporter, prof *profile.Profile, type_ string) []StackSample {
	typeIdx := -1
	for i, sampleType := range prof.SampleType {
		if sampleType.Type == type_ {
			typeIdx = i
		}
	}
	if typeIdx == -1 {
		r.Fatalf("Couldn't find sample type %s", type_)
	}

	if err := prof.Aggregate(true, true, false, false, false, false); err != nil {
		r.Fatalf("Error aggregating profile samples: %v", err)
	}
	prof = prof.Compact()
	sort.Slice(prof.Sample, func(i, j int) bool {
		return prof.Sample[i].Value[0] > prof.Sample[j].Value[0]
	})

	var out []StackSample
	for _, sample := range prof.Sample {
		var frames []string
		for i := range sample.Location {
			loc := sample.Location[len(sample.Location)-i-1]
			for j := range loc.Line {
				line := loc.Line[len(loc.Line)-j-1]
				name := line.Function.Name
				frames = append(frames, name)
			}
		}
		labels := make(map[string][]string)
		for k, v := range sample.Label {
			// ease the comparison by sorting string values
			sort.Strings(v)
			labels[k] = v
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

func checkLabels(r Reporter, labels map[string][]string, expectedLabels []Labels) bool {
	for _, expectedLabel := range expectedLabels {
		if values, ok := labels[expectedLabel.Key]; ok {
			if expectedLabel.Values != nil {
				// Right now all values should be present.
				if len(values) != len(expectedLabel.Values) {
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
						r.Fatalf("Error matching regexp %s: %v", v, err)
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

func assertStackWithFailureHandling(r Reporter, prof []StackSample, regexpStack string, valueOpt Optional[float64], pctOpt Optional[int64], epsilonPct int64, labels []Labels, allowFailure bool, hasFailures *bool) (matching int64) {
	rx, err := regexp.Compile(regexpStack)
	if err != nil {
		r.Fatalf("Error compiling regex: %v, %s", err, regexpStack)
	}
	var total int64 = 0
	for _, ss := range prof {
		total += ss.Val
		if rx.MatchString(ss.Stack) {
			if labels == nil || checkLabels(r, ss.Labels, labels) {
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
		if errorPct > float64(epsilonPct) {
			if allowFailure {
				r.Logf("\033[33mAssertion failed (allowed): stack '%s' (labels=%v) should have been %.1f +/- %d%% of the profile but was %d with %.1f%% error\033[0m", regexpStack, labels, value, epsilonPct, matching, errorPct)
				*hasFailures = true
			} else {
				r.Errorf("\033[31mAssertion failed: stack '%s' (labels=%v) should have been %.1f +/- %d%% of the profile but was %d with %.1f%% error\033[0m", regexpStack, labels, value, epsilonPct, matching, errorPct)
			}
		} else {
			r.Logf("\033[32mAssertion succeeded: stack '%s' (labels=%v) is %.1f +/- %d%% of the profile (was %d with %.1f%% error)\033[0m", regexpStack, labels, value, epsilonPct, matching, errorPct)
		}
	}

	if pct, ok := pctOpt.Value(); ok {
		diff := absDiff(pct, actualPct)
		if diff > epsilonPct {
			if allowFailure {
				r.Logf("\033[33mAssertion failed (allowed): stack '%s' (labels=%v) should have been %d%% +/- %d%% of the profile but was %d%% with %d%% error\033[0m", regexpStack, labels, pct, epsilonPct, actualPct, diff)
				*hasFailures = true
			} else {
				r.Errorf("\033[31mAssertion failed: stack '%s' (labels=%v) should have been %d%% +/- %d%% of the profile but was %d%% with %d%% error\033[0m", regexpStack, labels, pct, epsilonPct, actualPct, diff)
			}
		} else {
			r.Logf("\033[32mAssertion succeeded: stack '%s' (labels=%v) is %d%% +/- %d%% of the profile (was %d%% with %d%% error)\033[0m", regexpStack, labels, pct, epsilonPct, actualPct, diff)
		}
	}
	return
}

func analyzeProfDataWithFailureHandling(r Reporter, prof []StackSample, typedStacks TypedStacks, durationSecs float64, allowFailure bool) {
	var matchingSum int64 = 0
	var hasFailures bool = false

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

		matching := assertStackWithFailureHandling(r, prof, regexpStack, valueOpt, percent, errorMargin, stack.Labels, allowFailure, &hasFailures)
		matchingSum += matching
		// TODO: add an assertion on counts (e.g. number of allocations), not just summed values.
	}

	if expectedSum, ok := typedStacks.ValueMatchingSum.Value(); ok {
		value := float64(expectedSum)
		if durationSecs > 0 {
			// NOTE: When profile duration is bigger than 0, all values represent rates.
			value = value * durationSecs
		}
		errorPct := relDiff(float64(matchingSum), value)
		if errorPct > float64(typedStacks.ErrorMargin) {
			if allowFailure {
				r.Logf("\033[33mAssertion failed (allowed): profile '%s' should have total matching sum of %1.f +/- %d%% but was %d with %.1f%% error\033[0m", typedStacks.ProfileType, value, typedStacks.ErrorMargin, matchingSum, errorPct)
				hasFailures = true
			} else {
				r.Errorf("\033[31mAssertion failed: profile '%s' should have total matching sum of %1.f +/- %d%% but was %d with %.1f%% error\033[0m", typedStacks.ProfileType, value, typedStacks.ErrorMargin, matchingSum, errorPct)
			}
		} else {
			r.Logf("\033[32mAssertion succeeded: profile '%s' has total matching sum of %1.f +/- %d%% (was %d with %.1f%% error)\033[0m", typedStacks.ProfileType, value, typedStacks.ErrorMargin, matchingSum, errorPct)
		}
	}

	if allowFailure && hasFailures {
		r.Logf("\033[33mProfile analysis completed with failures (allowed for first profile)\033[0m")
	}
}

func writeToJSONFile(data StackTestData, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, 0644)
}

// ReadJSONFile loads, schema-validates and returns the expected_profile.json
// description at filePath.
func ReadJSONFile(filePath string) (StackTestData, error) {
	var data StackTestData
	byteValue, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	// Step 1: Validate JSON syntax
	if !json.Valid(byteValue) {
		return data, fmt.Errorf("invalid JSON syntax in %s", filePath)
	}

	// Step 2: Validate against schema
	schemaLoader := gojsonschema.NewStringLoader(expectedProfileSchema)
	documentLoader := gojsonschema.NewBytesLoader(byteValue)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return data, fmt.Errorf("schema validation error for %s: %v", filePath, err)
	}
	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return data, fmt.Errorf("JSON schema validation failed for %s:\n  - %s", filePath, strings.Join(errs, "\n  - "))
	}

	// Step 3: Unmarshal validated JSON
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return data, err
	}

	// Step 4: Validate rules
	if err := data.Validate(); err != nil {
		return data, fmt.Errorf("validation failed for %s: %v", filePath, err)
	}

	return data, nil
}

func getAllFiles(folder string) ([]string, error) {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
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

// ReadPprofFile reads a pprof file from disk, transparently decompressing lz4
// or zstd frames if present, and returns the parsed profile.
func ReadPprofFile(pprofFile string) (*profile.Profile, error) {
	file, err := os.Open(pprofFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
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

	// Handle zstd-compressed profiles.
	// RFC 8878 defines the zstd frame magic as little-endian 0xFD2FB528; the decoder expects this LE constant.
	if len(content) >= 4 {
		// parse the first 4 bytes as little-endian and compare to 0xFD2FB528
		magic := uint32(content[0]) | uint32(content[1])<<8 | uint32(content[2])<<16 | uint32(content[3])<<24
		if magic == 0xFD2FB528 {
			dec, err := zstd.NewReader(nil)
			if err != nil {
				return nil, err
			}
			decompressed, err := dec.DecodeAll(content, nil)
			dec.Close()
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

// AnalyzePprofFile reads a single pprof file and asserts the given typedStacks
// expectations against it. If captureData is true, a JSON dump of the actual
// stacks observed in the profile is written next to the pprof file (useful to
// bootstrap an expected_profile.json).
func AnalyzePprofFile(r Reporter, pprofFile string, typedStacks TypedStacks, testName string, captureData bool, scaleByDuration bool, allowFailure bool) {
	prof, err := ReadPprofFile(pprofFile)
	if err != nil {
		r.Fatalf("Error reading file %s", pprofFile)
	}
	r.Logf("Analyzing results in %s for profile type %s", pprofFile, typedStacks.ProfileType)

	profileDuration := float64(prof.DurationNanos) / 1000000000.0
	r.Logf("Found a profile duration of %.1f seconds (in %s)", profileDuration, filepath.Base(pprofFile))

	// Store current data in a json file to help users create their tests
	if captureData {
		captureProfData(r, prof, pprofFile, testName, profileDuration)
	}
	if !scaleByDuration {
		// ignore duration, values can be considered absolute
		profileDuration = 0
	}
	typedProf := getProfileType(r, prof, typedStacks.ProfileType)
	analyzeProfDataWithFailureHandling(r, typedProf, typedStacks, profileDuration, allowFailure)
}

// AnalyzeResults loads the expected_profile.json at jsonFilePath and asserts
// every pprof file under pprofFolder matches it. Failures are reported via r.
func AnalyzeResults(r Reporter, jsonFilePath string, pprofFolder string) {
	stackTestData, err := ReadJSONFile(jsonFilePath)
	if err != nil {
		r.Fatalf("Error opening file %s: %v", jsonFilePath, err)
	}

	var defaultPprofRegexp *regexp.Regexp
	if stackTestData.PprofRegex != "" {
		defaultPprofRegexp = regexp.MustCompile(stackTestData.PprofRegex)
	} else {
		// python files are in the form "profile.<pid>.number"
		// Other profilers (using pprof) include pprof in the name
		// Filter out files that ends with '.json' to avoid considering files dumped by captureProfData as profiles
		// Golang regexes do not have negative lookahed, so we need to use `([^n]|[^o]n|[^s]on|[^j]son|[^.]json)$` instead of `(?![.]json)$
		defaultPprofRegexp = regexp.MustCompile("^(profile|.*pprof)($|.*([^n]|[^o]n|[^s]on|[^j]son|[^.]json)$)")
	}
	processedProfilesMap := make(map[string]bool)

	for _, typedStacks := range stackTestData.Stacks {
		// use typedStack.PprofRegex if defined, otherwise use defaultPprofRegexp
		pprofRegexp := defaultPprofRegexp
		if typedStacks.PprofRegex != "" {
			pprofRegexp = regexp.MustCompile(typedStacks.PprofRegex)
		}
		matchingFiles, err := getMatchingFiles(pprofFolder, pprofRegexp)
		if err != nil {
			r.Fatalf("Error getting matching files: %v", err)
		}
		if len(matchingFiles) == 0 {
			r.Errorf("No matching files found for %s in %s", pprofRegexp, pprofFolder)

			if allFiles, err := getAllFiles(pprofFolder); err == nil {
				r.Errorf("All files: %v", allFiles)
			}
		} else {
			// Sort files by name to ensure consistent ordering
			sort.Strings(matchingFiles)

			for i, file := range matchingFiles {
				_, fileAlreadyProcessed := processedProfilesMap[file]
				if !fileAlreadyProcessed {
					processedProfilesMap[file] = true
				}

				// Allow failure for the first profile if the setting is enabled
				allowFailure := stackTestData.AllowFirstProfileFailure && i == 0
				if allowFailure {
					r.Logf("Analyzing first profile with failure tolerance enabled: %s", filepath.Base(file))
				}

				AnalyzePprofFile(r, file, typedStacks, stackTestData.TestName, !fileAlreadyProcessed, stackTestData.ScaleByDuration, allowFailure)
			}
		}
	}
}
