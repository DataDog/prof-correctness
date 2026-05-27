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

	"github.com/DataDog/profiler-correctness/v1/analysis"
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
