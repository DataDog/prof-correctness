package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testStack struct {
	Regex   string
	Percent int64
}

func writeCapture(t *testing.T, root, folder, testName, profileType string, stacks []testStack) {
	t.Helper()
	dir := filepath.Join(root, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := make([]stackEntry, len(stacks))
	for i, s := range stacks {
		p := s.Percent
		content[i] = stackEntry{RegularExpression: s.Regex, Percent: &p}
	}
	c := captureFile{
		TestName: testName,
		Stacks: []typedStacks{{
			ProfileType:  profileType,
			StackContent: content,
		}},
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestCompare_WithinBandPasses(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
		{Regex: "^cold$", Percent: 3},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 22},
		{Regex: "^cold$", Percent: 2},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 5, &stdout, &stderr); err != nil {
		t.Fatalf("run: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
}

func TestCompare_DivergenceInsideBandFails(t *testing.T) {
	// Both would pass a shared 20±5 band; 18 vs 24 is still a 6 pp drift.
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 18},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 24},
	})

	var stdout, stderr bytes.Buffer
	err := run(left, right, 5, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected divergence to fail")
	}
	if !strings.Contains(stderr.String(), "|18-24|=6 > 5") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCompare_MaxPPOverrideAllowsWiderDrift(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 18},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 24},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 10, &stdout, &stderr); err != nil {
		t.Fatalf("run: %v\nstderr=%s", err, stderr.String())
	}
}

func TestCompare_UnmatchedSignificantFails(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
		{Regex: "^only14$", Percent: 8},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
	})

	var stdout, stderr bytes.Buffer
	err := run(left, right, 5, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unmatched ≥5% stack to fail")
	}
	if !strings.Contains(stderr.String(), "only14") || !strings.Contains(stderr.String(), "unmatched") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCompare_UnmatchedTailIgnored(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
		{Regex: "^noise$", Percent: 3},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 5, &stdout, &stderr); err != nil {
		t.Fatalf("tail should be ignored: %v\nstderr=%s", err, stderr.String())
	}
}

func TestCompare_PairsByTestNameWhenFolderHasNoVersion(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "run-left", "python_alloc_3.14", "alloc-space", []testStack{
		{Regex: "^alloc$", Percent: 40},
	})
	writeCapture(t, right, "run-right", "python_alloc_3.15", "alloc-space", []testStack{
		{Regex: "^alloc$", Percent: 41},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 5, &stdout, &stderr); err != nil {
		t.Fatalf("test_name pairing: %v\nstderr=%s", err, stderr.String())
	}
}

func TestCompare_MissingFamilyFails(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
	})
	writeCapture(t, right, "python_alloc_3.15-20260101-120100-bbb", "python_alloc", "alloc-space", []testStack{
		{Regex: "^alloc$", Percent: 40},
	})

	var stdout, stderr bytes.Buffer
	err := run(left, right, 5, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected missing family to fail")
	}
	if !strings.Contains(stderr.String(), "missing on right") && !strings.Contains(stderr.String(), "missing on left") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCompare_NestedDataDir(t *testing.T) {
	// download-artifact may preserve a data/ prefix depending on upload path.
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, filepath.Join("data", "python_lock_3.14-20260101-120000-aaa"), "python_lock", "lock-acquire", []testStack{
		{Regex: "^lock$", Percent: 50},
	})
	writeCapture(t, right, filepath.Join("data", "python_lock_3.15-20260101-120100-bbb"), "python_lock", "lock-acquire", []testStack{
		{Regex: "^lock$", Percent: 52},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 5, &stdout, &stderr); err != nil {
		t.Fatalf("nested data/: %v\nstderr=%s", err, stderr.String())
	}
}

func TestCompare_LastJSONWins(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	dir := filepath.Join(left, "python_cpu_3.14-20260101-120000-aaa")
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 10},
	})
	// Later pprof capture (sorted name) should override the first.
	later := captureFile{
		TestName: "python_cpu",
		Stacks: []typedStacks{{
			ProfileType: "cpu-time",
			StackContent: []stackEntry{{
				RegularExpression: "^hot$",
				Percent:           ptr(int64(20)),
			}},
		}},
	}
	raw, err := json.Marshal(later)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "z-later.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 5, &stdout, &stderr); err != nil {
		t.Fatalf("last json should win (20 vs 20): %v\nstderr=%s", err, stderr.String())
	}
}

func TestCompare_EmptyDirFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(t.TempDir(), t.TempDir(), 5, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected empty dirs to fail")
	}
	if !strings.Contains(err.Error(), "no capture JSON") {
		t.Fatalf("err=%v", err)
	}
}

func TestFamilyFromName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"python_cpu_3.14-20260101-120000-aaa", "python_cpu"},
		{"python_live_heap_3.15-20260101-120000-bbb", "python_live_heap"},
		{"python_cpu", ""},
		{"python_alloc_3.14", "python_alloc"},
	}
	for _, c := range cases {
		if got := familyFromName(c.in); got != c.want {
			t.Errorf("familyFromName(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func writeAsserted(t *testing.T, scenariosDir, family, profileType, regex string) {
	t.Helper()
	dir := filepath.Join(scenariosDir, family)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	c := captureFile{
		TestName: family,
		Stacks: []typedStacks{{
			ProfileType: profileType,
			StackContent: []stackEntry{{
				RegularExpression: regex,
			}},
		}},
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestCompare_FactorialAliasPairs(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_native_cpu_3.14-20260101-120000-aaa", "python_native_cpu", "cpu-time", []testStack{
		{Regex: `^<module>;main;factorial_work;math\.factorial$`, Percent: 21},
	})
	writeCapture(t, right, "python_native_cpu_3.15-20260101-120100-bbb", "python_native_cpu", "cpu-time", []testStack{
		{Regex: `^<module>;main;factorial_work;math\.integer\.factorial$`, Percent: 20},
	})

	var stdout, stderr bytes.Buffer
	if err := run(left, right, 5, &stdout, &stderr); err != nil {
		t.Fatalf("alias should pair: %v\nstderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "3.14=21 3.15=20") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestCompare_AssertedOnlyIgnoresExtra(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	scenarios := t.TempDir()
	writeAsserted(t, scenarios, "python_cpu", "cpu-time", "^hot$")
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
		{Regex: "^CodeProvenance$", Percent: 40},
		{Regex: ".*sleep$", Percent: 30},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 22},
		{Regex: "^pathlib$", Percent: 50},
	})

	var stdout, stderr bytes.Buffer
	err := runOpts(runConfig{
		leftDir: left, rightDir: right, maxPP: 5,
		scenariosDir: scenarios, stdout: &stdout, stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("extra capture stacks should be ignored: %v\nstderr=%s", err, stderr.String())
	}
	if strings.Contains(stdout.String(), "CodeProvenance") || strings.Contains(stderr.String(), "unmatched") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCompare_Exceptions6ppStillFails(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_exceptions_3.14-20260101-120000-aaa", "python_exceptions", "cpu-time", []testStack{
		{Regex: `^<module>;main;handle_value_error$`, Percent: 41},
	})
	writeCapture(t, right, "python_exceptions_3.15-20260101-120100-bbb", "python_exceptions", "cpu-time", []testStack{
		{Regex: `^<module>;main;handle_value_error$`, Percent: 47},
	})

	var stdout, stderr bytes.Buffer
	err := run(left, right, 5, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected 6pp to fail at -max-pp 5")
	}
	if !strings.Contains(stderr.String(), "|41-47|=6 > 5") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCompare_ExcludeSkipsDivergence(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	writeCapture(t, left, "python_exceptions_3.14-20260101-120000-aaa", "python_exceptions", "cpu-time", []testStack{
		{Regex: `^<module>;main;handle_value_error$`, Percent: 41},
	})
	writeCapture(t, right, "python_exceptions_3.15-20260101-120100-bbb", "python_exceptions", "cpu-time", []testStack{
		{Regex: `^<module>;main;handle_value_error$`, Percent: 47},
	})
	writeCapture(t, left, "python_live_heap_3.14-20260101-120000-aaa", "python_live_heap", "heap-live-samples", []testStack{
		{Regex: `^<module>;main;Target\.run;Target\.retain_major$`, Percent: 64},
	})
	writeCapture(t, right, "python_live_heap_3.15-20260101-120100-bbb", "python_live_heap", "heap-live-samples", []testStack{
		{Regex: `^<module>;main;Target\.run;Target\.retain_major$`, Percent: 79},
	})
	writeCapture(t, left, "python_cpu_3.14-20260101-120000-aaa", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 20},
	})
	writeCapture(t, right, "python_cpu_3.15-20260101-120100-bbb", "python_cpu", "cpu-time", []testStack{
		{Regex: "^hot$", Percent: 21},
	})

	var stdout, stderr bytes.Buffer
	err := runOpts(runConfig{
		leftDir: left, rightDir: right, maxPP: 5,
		exclude: parseExclude("python_exceptions,python_live_heap"),
		stdout:  &stdout, stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("excluded families should not fail: %v\nstderr=%s", err, stderr.String())
	}
	if strings.Contains(stderr.String(), "handle_value_error") || strings.Contains(stderr.String(), "retain_major") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCompare_ExcludeShortName(t *testing.T) {
	if !isExcluded("python_exceptions", parseExclude("exceptions,live_heap")) {
		t.Fatal("exceptions should match python_exceptions")
	}
	if !isExcluded("python_live_heap", parseExclude("exceptions,live_heap")) {
		t.Fatal("live_heap should match python_live_heap")
	}
	if isExcluded("python_cpu", parseExclude("exceptions,live_heap")) {
		t.Fatal("python_cpu should not be excluded")
	}
}

func TestNormalizeRegex_Factorial(t *testing.T) {
	got := normalizeRegex(`^<module>;main;factorial_work;math\.integer\.factorial$`)
	want := `^<module>;main;factorial_work;math\.factorial$`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if normalizeRegex(want) != want {
		t.Fatal("already-canonical key should be unchanged")
	}
}

func ptr(v int64) *int64 { return &v }
