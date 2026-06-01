// External test (package analysis_test) — only exported names are visible, so
// this file doubles as a compile-time check that the public API surface
// consumed by external repos (e.g. dd-win-prof) keeps working.
package analysis_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/pprof/profile"

	"github.com/DataDog/prof-correctness/analysis"
)

// writeMinimalPprof writes a tiny but well-formed cpu-time pprof to dir. The
// profile has a single sample worth wantNanos in a function named fnName.
func writeMinimalPprof(t *testing.T, dir, fnName string, wantNanos int64) string {
	t.Helper()
	fn := &profile.Function{ID: 1, Name: fnName}
	loc := &profile.Location{ID: 1, Line: []profile.Line{{Function: fn}}}
	p := &profile.Profile{
		SampleType: []*profile.ValueType{{Type: "cpu-time", Unit: "nanoseconds"}},
		PeriodType: &profile.ValueType{Type: "cpu-time", Unit: "nanoseconds"},
		Period:     10_000_000,
		Sample: []*profile.Sample{
			{Value: []int64{wantNanos}, Location: []*profile.Location{loc}},
		},
		Function: []*profile.Function{fn},
		Location: []*profile.Location{loc},
	}
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("write pprof: %v", err)
	}
	path := filepath.Join(dir, "profile.pprof")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

func writeJSON(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "expected.json")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	return path
}

// TestPublicAPI_HappyPath drives AnalyzeResults through StdReporter + Run the
// way an external consumer (e.g. dd-win-prof) will.
func TestPublicAPI_HappyPath(t *testing.T) {
	dir := t.TempDir()
	writeMinimalPprof(t, dir, "hot_function", 10_000_000)
	jsonPath := writeJSON(t, dir, `{
		"test_name": "api-happy",
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [
				{"regular_expression": "^hot_function$", "value": 10000000, "error_margin": 0}
			]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if r.Failed() {
		t.Fatal("expected Failed()=false on a matching profile")
	}
}

// TestPublicAPI_FailingAssertion confirms Failed() is set when the profile
// doesn't match expectations (Errorf path, no Fatalf).
func TestPublicAPI_FailingAssertion(t *testing.T) {
	dir := t.TempDir()
	writeMinimalPprof(t, dir, "hot_function", 10_000_000)
	jsonPath := writeJSON(t, dir, `{
		"test_name": "api-fail",
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [
				{"regular_expression": "^hot_function$", "value": 99999999, "error_margin": 0}
			]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if !r.Failed() {
		t.Fatal("expected Failed()=true on a value mismatch")
	}
}

// TestPublicAPI_MinMaxValue_Pass confirms min_value/max_value bounds pass when
// the matched value sits within the bounds.
func TestPublicAPI_MinMaxValue_Pass(t *testing.T) {
	dir := t.TempDir()
	writeMinimalPprof(t, dir, "hot_function", 10_000_000)
	jsonPath := writeJSON(t, dir, `{
		"test_name": "api-minmax-pass",
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [
				{"regular_expression": "^hot_function$", "min_value": 5000000, "max_value": 20000000}
			]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if r.Failed() {
		t.Fatal("expected Failed()=false when value is within [min_value, max_value]")
	}
}

// TestPublicAPI_MinValue_Fail confirms a min_value bound fails when the matched
// value falls below the floor (the assertion idle_baseline relies on).
func TestPublicAPI_MinValue_Fail(t *testing.T) {
	dir := t.TempDir()
	writeMinimalPprof(t, dir, "hot_function", 10_000_000)
	jsonPath := writeJSON(t, dir, `{
		"test_name": "api-min-fail",
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [
				{"regular_expression": "^hot_function$", "min_value": 20000000}
			]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if !r.Failed() {
		t.Fatal("expected Failed()=true when value is below min_value")
	}
}

// TestPublicAPI_FatalfRecovers confirms Run() catches Fatalf panics so the
// caller can inspect Failed() and exit normally — instead of the panic
// propagating out of the library.
func TestPublicAPI_FatalfRecovers(t *testing.T) {
	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, "/does/not/exist.json", t.TempDir())
	})
	if !r.Failed() {
		t.Fatal("expected Failed()=true after Fatalf on missing expected_profile.json")
	}
}

// TestPublicAPI_ReadPprofFile confirms the lz4/zstd/raw pprof loader is part
// of the public surface and round-trips a written profile.
func TestPublicAPI_ReadPprofFile(t *testing.T) {
	dir := t.TempDir()
	path := writeMinimalPprof(t, dir, "fn_a", 42)
	prof, err := analysis.ReadPprofFile(path)
	if err != nil {
		t.Fatalf("ReadPprofFile: %v", err)
	}
	if len(prof.SampleType) != 1 || prof.SampleType[0].Type != "cpu-time" {
		t.Fatalf("unexpected SampleType: %+v", prof.SampleType)
	}
	if len(prof.Sample) != 1 || prof.Sample[0].Value[0] != 42 {
		t.Fatalf("unexpected sample value: %+v", prof.Sample)
	}
}

// TestPublicAPI_ReadJSONFile confirms the schema validator is reachable as a
// public function (consumers may want to validate expected_profile.json files
// independently of running an analysis).
func TestPublicAPI_ReadJSONFile(t *testing.T) {
	dir := t.TempDir()
	good := writeJSON(t, dir, `{
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [{"regular_expression": "^x$", "value": 1}]
		}]
	}`)
	if _, err := analysis.ReadJSONFile(good); err != nil {
		t.Fatalf("ReadJSONFile rejected a valid file: %v", err)
	}

	badPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badPath, []byte(`{"stacks": []}`), 0644); err != nil {
		t.Fatalf("write bad: %v", err)
	}
	if _, err := analysis.ReadJSONFile(badPath); err == nil {
		t.Fatal("ReadJSONFile accepted an empty-stacks file without a note")
	}
}

// Compile-time check: *testing.T must satisfy analysis.Reporter so consumers
// can pass `t` straight through (the existing test files in package main rely
// on this — locking it in here protects external consumers too).
var _ analysis.Reporter = (*testing.T)(nil)

// writeManyStacksPprof writes a pprof with N distinct stacks, one sample each
// (different value per sample). Used to verify captureProfData captures every
// stack, including ones whose individual share is well under 1%.
func writeManyStacksPprof(t *testing.T, dir string, nStacks int) string {
	t.Helper()
	p := &profile.Profile{
		SampleType: []*profile.ValueType{{Type: "cpu-time", Unit: "nanoseconds"}},
		PeriodType: &profile.ValueType{Type: "cpu-time", Unit: "nanoseconds"},
		Period:     10_000_000,
	}
	for i := 0; i < nStacks; i++ {
		fn := &profile.Function{ID: uint64(i + 1), Name: "fn_" + strconv.Itoa(i)}
		loc := &profile.Location{ID: uint64(i + 1), Line: []profile.Line{{Function: fn}}}
		p.Function = append(p.Function, fn)
		p.Location = append(p.Location, loc)
		p.Sample = append(p.Sample, &profile.Sample{
			Value:    []int64{int64(i + 1)},
			Location: []*profile.Location{loc},
		})
	}
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("write pprof: %v", err)
	}
	path := filepath.Join(dir, "profile.pprof")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

// TestCaptureProfData_KeepsLongTail confirms captureProfData no longer drops
// stacks below 1% of the profile. We feed in 200 distinct stacks (each ~0.5%)
// and expect every one to appear in the captured JSON.
func TestCaptureProfData_KeepsLongTail(t *testing.T) {
	const nStacks = 200
	dir := t.TempDir()
	writeManyStacksPprof(t, dir, nStacks)
	jsonPath := writeJSON(t, dir, `{
		"test_name": "capture-keeps-long-tail",
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [{"regular_expression": ".*", "percent": 100, "error_margin": 100}]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if r.Failed() {
		t.Fatal("analyzer reported failure on permissive expected file")
	}

	// captureProfData drops the JSON next to the pprof, with the pprof's last
	// extension replaced by .json.
	captured := filepath.Join(dir, "profile.json")
	raw, err := os.ReadFile(captured)
	if err != nil {
		t.Fatalf("read captured json: %v", err)
	}

	var captureData struct {
		Stacks []struct {
			ProfileType  string `json:"profile-type"`
			StackContent []any  `json:"stack-content"`
		} `json:"stacks"`
	}
	if err := json.Unmarshal(raw, &captureData); err != nil {
		t.Fatalf("unmarshal captured json: %v", err)
	}

	var got int
	for _, s := range captureData.Stacks {
		if s.ProfileType == "cpu-time" {
			got = len(s.StackContent)
		}
	}
	if got != nStacks {
		t.Errorf("expected %d captured stacks (no filter), got %d", nStacks, got)
	}
	// Spot-check the lowest-percentage stack (fn_0, val=1) is present.
	if !strings.Contains(string(raw), "fn_0") {
		t.Error("expected fn_0 (smallest stack) in captured JSON; it was dropped")
	}
}

// writeRepeatedStackWithLabelsPprof writes a pprof where the SAME stack
// appears N times, with the same stable label (thread_name) but every sample
// carrying a different ephemeral label (end_timestamp_ns). After capture and
// label-stripping, all N samples should collapse into a single entry whose
// value is the sum.
func writeRepeatedStackWithLabelsPprof(t *testing.T, dir string, nSamples int) string {
	t.Helper()
	fn := &profile.Function{ID: 1, Name: "hot"}
	loc := &profile.Location{ID: 1, Line: []profile.Line{{Function: fn}}}
	p := &profile.Profile{
		SampleType: []*profile.ValueType{{Type: "cpu-time", Unit: "nanoseconds"}},
		PeriodType: &profile.ValueType{Type: "cpu-time", Unit: "nanoseconds"},
		Period:     10_000_000,
		Function:   []*profile.Function{fn},
		Location:   []*profile.Location{loc},
	}
	for i := 0; i < nSamples; i++ {
		p.Sample = append(p.Sample, &profile.Sample{
			Value:    []int64{int64(i + 1)},
			Location: []*profile.Location{loc},
			Label: map[string][]string{
				"thread_name": {"worker"},
			},
			NumLabel: map[string][]int64{
				"end_timestamp_ns": {int64(1_000_000_000 + i)},
				"thread id":        {int64(100 + i)},
			},
		})
	}
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("write pprof: %v", err)
	}
	path := filepath.Join(dir, "profile.pprof")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

// TestCaptureProfData_GroupsByStackAndStableLabels confirms ephemeral labels
// (end_timestamp_ns, thread id, process_id, …) are stripped and samples
// sharing the same (stack, stable-labels) signature collapse into one entry
// whose value is the sum of the source samples.
func TestCaptureProfData_GroupsByStackAndStableLabels(t *testing.T) {
	const nSamples = 50
	dir := t.TempDir()
	writeRepeatedStackWithLabelsPprof(t, dir, nSamples)
	jsonPath := writeJSON(t, dir, `{
		"test_name": "capture-groups",
		"stacks": [{
			"profile-type": "cpu-time",
			"stack-content": [{"regular_expression": ".*", "percent": 100, "error_margin": 100}]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if r.Failed() {
		t.Fatal("analyzer reported failure on permissive expected file")
	}

	raw, err := os.ReadFile(filepath.Join(dir, "profile.json"))
	if err != nil {
		t.Fatalf("read captured json: %v", err)
	}

	var captured struct {
		Stacks []struct {
			ProfileType  string `json:"profile-type"`
			StackContent []struct {
				Value  int64 `json:"value"`
				Labels []struct {
					Key    string   `json:"key"`
					Values []string `json:"values"`
				} `json:"labels"`
			} `json:"stack-content"`
		} `json:"stacks"`
	}
	if err := json.Unmarshal(raw, &captured); err != nil {
		t.Fatalf("unmarshal captured json: %v", err)
	}

	var cpuStack *struct {
		ProfileType  string `json:"profile-type"`
		StackContent []struct {
			Value  int64 `json:"value"`
			Labels []struct {
				Key    string   `json:"key"`
				Values []string `json:"values"`
			} `json:"labels"`
		} `json:"stack-content"`
	}
	for i := range captured.Stacks {
		if captured.Stacks[i].ProfileType == "cpu-time" {
			cpuStack = &captured.Stacks[i]
			break
		}
	}
	if cpuStack == nil {
		t.Fatal("cpu-time stack missing from capture")
	}

	if got := len(cpuStack.StackContent); got != 1 {
		t.Fatalf("expected 1 grouped entry (same stack + same stable labels), got %d", got)
	}

	want := int64(nSamples * (nSamples + 1) / 2) // 1+2+…+nSamples
	if got := cpuStack.StackContent[0].Value; got != want {
		t.Errorf("expected summed value %d, got %d", want, got)
	}

	// Labels should contain thread_name=worker and nothing else.
	gotLabels := cpuStack.StackContent[0].Labels
	if len(gotLabels) != 1 {
		t.Fatalf("expected exactly one kept label (thread_name), got %d: %+v", len(gotLabels), gotLabels)
	}
	if gotLabels[0].Key != "thread_name" || len(gotLabels[0].Values) != 1 || gotLabels[0].Values[0] != "worker" {
		t.Errorf("expected thread_name=worker, got %+v", gotLabels[0])
	}

	// Sanity: ephemeral keys should NOT appear in the captured JSON.
	for _, banned := range []string{"end_timestamp_ns", "thread id", "process_id", "span id", "local root span id"} {
		if strings.Contains(string(raw), `"key": "`+banned+`"`) {
			t.Errorf("ephemeral label %q leaked into captured JSON", banned)
		}
	}
}

// TestCaptureProfData_LowCountRateNotTruncated guards against per-sample
// rate scaling: two samples of value 1 over a 2 s profile must group to a
// rate of 1, not 0. Scaling each sample individually (int64(1.0/2.0)=0)
// before summing would truncate to 0; raw values must be summed first.
func TestCaptureProfData_LowCountRateNotTruncated(t *testing.T) {
	dir := t.TempDir()

	fn := &profile.Function{ID: 1, Name: "rare"}
	loc := &profile.Location{ID: 1, Line: []profile.Line{{Function: fn}}}
	p := &profile.Profile{
		SampleType:    []*profile.ValueType{{Type: "cpu-samples", Unit: "count"}},
		PeriodType:    &profile.ValueType{Type: "cpu-time", Unit: "nanoseconds"},
		Period:        10_000_000,
		DurationNanos: 2_000_000_000, // 2 seconds — triggers rate scaling
		Function:      []*profile.Function{fn},
		Location:      []*profile.Location{loc},
		Sample: []*profile.Sample{
			{Value: []int64{1}, Location: []*profile.Location{loc}},
			{Value: []int64{1}, Location: []*profile.Location{loc}},
		},
	}
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("write pprof: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "profile.pprof"), buf.Bytes(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	jsonPath := writeJSON(t, dir, `{
		"test_name": "low-count-rate",
		"scale_by_duration": true,
		"stacks": [{
			"profile-type": "cpu-samples",
			"stack-content": [{"regular_expression": ".*", "percent": 100, "error_margin": 100}]
		}]
	}`)

	r := analysis.NewStdReporter(os.Stdout, os.Stderr)
	analysis.Run(r, func() {
		analysis.AnalyzeResults(r, jsonPath, dir)
	})
	if r.Failed() {
		t.Fatal("analyzer reported failure on permissive expected file")
	}

	raw, err := os.ReadFile(filepath.Join(dir, "profile.json"))
	if err != nil {
		t.Fatalf("read captured json: %v", err)
	}
	var captured struct {
		Stacks []struct {
			ProfileType  string `json:"profile-type"`
			StackContent []struct {
				Value int64 `json:"value"`
			} `json:"stack-content"`
		} `json:"stacks"`
	}
	if err := json.Unmarshal(raw, &captured); err != nil {
		t.Fatalf("unmarshal captured json: %v", err)
	}

	if len(captured.Stacks) != 1 || len(captured.Stacks[0].StackContent) != 1 {
		t.Fatalf("expected exactly one captured entry, got %+v", captured)
	}
	// 2 raw samples summed = 2, scaled by 2s duration = 1 sample/sec.
	if got := captured.Stacks[0].StackContent[0].Value; got != 1 {
		t.Errorf("expected grouped rate value=1 (2 samples / 2s), got %d — pre-grouping scaling truncated", got)
	}
}
