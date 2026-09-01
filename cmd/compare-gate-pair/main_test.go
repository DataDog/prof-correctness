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

func writeJSON(t *testing.T, dir, file, testName, profileType string, stacks []stackEntry) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	raw, err := json.Marshal(captureFile{
		TestName: testName,
		Stacks:   []typedStacks{{ProfileType: profileType, StackContent: stacks}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func writeCapture(t *testing.T, root, folder, profileType string, stacks []testStack) {
	t.Helper()
	content := make([]stackEntry, len(stacks))
	for i, s := range stacks {
		p := s.Percent
		content[i] = stackEntry{RegularExpression: s.Regex, Percent: &p}
	}
	writeJSON(t, filepath.Join(root, folder), "profile.json", folder, profileType, content)
}

func writeAsserted(t *testing.T, scenariosDir, family, profileType, regex string) {
	t.Helper()
	writeJSON(t, filepath.Join(scenariosDir, family+"_3.14"), "expected_profile.json", family, profileType,
		[]stackEntry{{RegularExpression: regex}})
}

func cmpRun(t *testing.T, left, right, scenarios string, exclude []string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := run(runConfig{
		leftDir: left, rightDir: right, maxPP: 5, scenariosDir: scenarios,
		exclude: exclude, stdout: &stdout, stderr: &stderr,
	})
	return stdout.String(), stderr.String(), err
}

func TestCompare_DivergenceFails(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 18}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 24}})
	_, stderr, err := cmpRun(t, left, right, "", nil)
	if err == nil || !strings.Contains(stderr, "|18-24|=6 > 5") {
		t.Fatalf("expected divergence: err=%v stderr=%q", err, stderr)
	}
}

func TestCompare_Unmatched(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}, {"^only14$", 8}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 20}})
	_, stderr, err := cmpRun(t, left, right, "", nil)
	if err == nil || !strings.Contains(stderr, "only14") || !strings.Contains(stderr, "unmatched") {
		t.Fatalf("expected unmatched ≥5%% fail: err=%v stderr=%q", err, stderr)
	}

	left2, right2 := t.TempDir(), t.TempDir()
	writeCapture(t, left2, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}, {"^noise$", 3}})
	writeCapture(t, right2, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 20}})
	if _, stderr, err := cmpRun(t, left2, right2, "", nil); err != nil {
		t.Fatalf("tail <5%% should be ignored: %v\nstderr=%s", err, stderr)
	}
}

func TestCompare_FactorialAlias(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "native_3.14", "cpu-time", []testStack{{`^math\.factorial$`, 21}})
	writeCapture(t, right, "native_3.15", "cpu-time", []testStack{{`^math\.integer\.factorial$`, 20}})
	stdout, stderr, err := cmpRun(t, left, right, "", nil)
	if err != nil {
		t.Fatalf("alias should pair: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "3.14=21 3.15=20") {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCompare_AnchoredFold(t *testing.T) {
	left, right, scenarios := t.TempDir(), t.TempDir(), t.TempDir()
	writeAsserted(t, scenarios, "cpu", "cpu-time", `^.*Foo\.b$`)
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{`^x;Foo.b$`, 10}, {`^y;Foo\.b$`, 10}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{`^x;Foo.b$`, 11}, {`^y;Foo\.b$`, 10}})
	stdout, stderr, err := cmpRun(t, left, right, scenarios, nil)
	if err != nil {
		t.Fatalf("anchored asserted key should fold capture body: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "3.14=20 3.15=21") {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCompare_QuoteMetaUnescape(t *testing.T) {
	left, right, scenarios := t.TempDir(), t.TempDir(), t.TempDir()
	writeAsserted(t, scenarios, "cpu", "cpu-time", `^.*foo\+bar$`)
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{`^x;foo\+bar$`, 10}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{`^x;foo\+bar$`, 11}})
	stdout, stderr, err := cmpRun(t, left, right, scenarios, nil)
	if err != nil {
		t.Fatalf("QuoteMeta \\+ should unescape: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "3.14=10 3.15=11") {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCompare_EmptyFoldFails(t *testing.T) {
	left, right, scenarios := t.TempDir(), t.TempDir(), t.TempDir()
	writeAsserted(t, scenarios, "cpu", "cpu-time", `^Foo$`)
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^x$", 20}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^x$", 20}})
	_, stderr, err := cmpRun(t, left, right, scenarios, nil)
	if err == nil || !strings.Contains(stderr, "folded to nothing") {
		t.Fatalf("expected empty-fold fail: err=%v stderr=%q", err, stderr)
	}
	if strings.Contains(stderr, "|") || strings.Contains(stderr, "unmatched") {
		t.Fatalf("empty fold is a regex miss, not a percent miss: stderr=%q", stderr)
	}
}

func TestCompare_AssertedOnly(t *testing.T) {
	left, right, scenarios := t.TempDir(), t.TempDir(), t.TempDir()
	writeAsserted(t, scenarios, "cpu", "cpu-time", ".*hot")
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}, {"^extra$", 40}, {".*sleep$", 30}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 22}, {"^other$", 50}})
	stdout, stderr, err := cmpRun(t, left, right, scenarios, nil)
	if err != nil {
		t.Fatalf("extra capture stacks should be ignored: %v\nstderr=%s", err, stderr)
	}
	if strings.Contains(stdout, "extra") || strings.Contains(stderr, "unmatched") {
		t.Fatalf("stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCompare_Exclude(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_exceptions_3.14", "cpu-time", []testStack{{"^e$", 41}})
	writeCapture(t, right, "python_exceptions_3.15", "cpu-time", []testStack{{"^e$", 47}})
	writeCapture(t, left, "python_live_heap_3.14", "heap-live-samples", []testStack{{"^h$", 64}})
	writeCapture(t, right, "python_live_heap_3.15", "heap-live-samples", []testStack{{"^h$", 79}})
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 21}})
	_, stderr, err := cmpRun(t, left, right, "", parseExclude("python_exceptions,python_live_heap"))
	if err != nil {
		t.Fatalf("excluded families should not fail: %v\nstderr=%s", err, stderr)
	}
	if strings.Contains(stderr, "python_exceptions") || strings.Contains(stderr, "python_live_heap") {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestCompare_MissingFamily(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "alloc_3.15", "alloc-space", []testStack{{"^a$", 40}})
	_, stderr, err := cmpRun(t, left, right, "", nil)
	if err == nil || (!strings.Contains(stderr, "missing on right") && !strings.Contains(stderr, "missing on left")) {
		t.Fatalf("expected missing family: err=%v stderr=%q", err, stderr)
	}
}

func TestFamilyFromName(t *testing.T) {
	if got := familyFromName("python_cpu_3.14-x"); got != "python_cpu" {
		t.Fatalf("got %q", got)
	}
	if got := familyFromName("python_live_heap_3.15-x"); got != "python_live_heap" {
		t.Fatalf("got %q", got)
	}
	if familyFromName("python_cpu") != "" {
		t.Fatal("unversioned name should not strip")
	}
	if got := familyOf("left/data/python_lock_3.14-ts/profile.json", ""); got != "python_lock" {
		t.Fatalf("nested data/: got %q", got)
	}
}

func TestCompare_EmptyDir(t *testing.T) {
	_, _, err := cmpRun(t, t.TempDir(), t.TempDir(), "", nil)
	if err == nil || !strings.Contains(err.Error(), "no capture JSON") {
		t.Fatalf("err=%v", err)
	}
}
