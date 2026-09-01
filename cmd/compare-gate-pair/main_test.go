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

func ptr(v int64) *int64 { return &v }
