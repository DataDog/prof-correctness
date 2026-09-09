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
	writeAsserted(t, scenarios, "cpu", "cpu-time", `^.*foo\+bar$`)
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{`^x;foo+bar$`, 10}, {`^y;foo\+bar$`, 10}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{`^x;foo+bar$`, 11}, {`^y;foo\+bar$`, 10}})
	stdout, stderr, err := cmpRun(t, left, right, scenarios, nil)
	if err != nil {
		t.Fatalf("anchored fold should unescape QuoteMeta \\+ and sum bodies: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "3.14=20 3.15=21") {
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
	writeCapture(t, left, "python_live_heap_3.14", "heap-live-samples", []testStack{{"^h$", 64}})
	writeCapture(t, right, "python_live_heap_3.15", "heap-live-samples", []testStack{{"^h$", 79}})
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 21}})
	_, stderr, err := cmpRun(t, left, right, "", parseExclude("python_live_heap"))
	if err != nil {
		t.Fatalf("excluded families should not fail: %v\nstderr=%s", err, stderr)
	}
	if strings.Contains(stderr, "python_live_heap") {
		t.Fatalf("stderr=%q", stderr)
	}
}

func TestCompare_ExcludeSkipsDuplicateDumps(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	p64 := int64(64)
	stacks := []stackEntry{{RegularExpression: "^h$", Percent: &p64}}
	writeJSON(t, filepath.Join(left, "python_live_heap_3.14"), "profiles.7.1.json", "python_live_heap_3.14", "heap-live-samples", stacks)
	writeJSON(t, filepath.Join(left, "python_live_heap_3.14"), "profiles.7.2.json", "python_live_heap_3.14", "heap-live-samples", stacks)
	writeJSON(t, filepath.Join(right, "python_live_heap_3.15"), "profiles.7.1.json", "python_live_heap_3.15", "heap-live-samples", stacks)
	writeJSON(t, filepath.Join(right, "python_live_heap_3.15"), "profiles.7.2.json", "python_live_heap_3.15", "heap-live-samples", stacks)
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 21}})
	stdout, stderr, err := cmpRun(t, left, right, "", parseExclude("python_live_heap"))
	if err != nil || strings.Contains(stdout+stderr, "python_live_heap") || !strings.Contains(stdout, "cpu") {
		t.Fatalf("excluded two-dump family must be skipped: %v stdout=%q stderr=%q", err, stdout, stderr)
	}
}

func TestCompare_SameRegexDifferentLabels(t *testing.T) {
	left, right, scenarios := t.TempDir(), t.TempDir(), t.TempDir()
	short := []labelSpec{{Key: "task name", Values: []string{"short_task"}}}
	long := []labelSpec{{Key: "task name", Values: []string{"long_task"}}}
	p33, p67, p34, p66 := int64(33), int64(67), int64(34), int64(66)
	writeJSON(t, filepath.Join(scenarios, "aio_3.14"), "expected_profile.json", "aio", "wall-time",
		[]stackEntry{
			{RegularExpression: ".*my_coroutine.*", Labels: short},
			{RegularExpression: ".*my_coroutine.*", Labels: long},
		})
	writeJSON(t, filepath.Join(left, "aio_3.14"), "profile.json", "aio_3.14", "wall-time",
		[]stackEntry{
			{RegularExpression: "^x;my_coroutine$", Percent: &p33, Labels: short},
			{RegularExpression: "^x;my_coroutine$", Percent: &p67, Labels: long},
		})
	writeJSON(t, filepath.Join(right, "aio_3.15"), "profile.json", "aio_3.15", "wall-time",
		[]stackEntry{
			{RegularExpression: "^x;my_coroutine$", Percent: &p34, Labels: short},
			{RegularExpression: "^x;my_coroutine$", Percent: &p66, Labels: long},
		})
	stdout, stderr, err := cmpRun(t, left, right, scenarios, nil)
	if err != nil {
		t.Fatalf("labeled rows should stay two keys: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "3.14=33 3.15=34") || !strings.Contains(stdout, "3.14=67 3.15=66") {
		t.Fatalf("expected two keys, got stdout=%q", stdout)
	}
	if strings.Contains(stdout, "3.14=100") {
		t.Fatalf("collapsed same-regex labels: stdout=%q", stdout)
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

func TestCompare_PythonCaptureNames(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	p20, p21 := int64(20), int64(21)
	writeJSON(t, filepath.Join(left, "cpu_3.14"), "profiles.7.1.json", "cpu_3.14", "cpu-time",
		[]stackEntry{{RegularExpression: "^hot$", Percent: &p20}})
	writeJSON(t, filepath.Join(right, "cpu_3.15"), "profile.9.json", "cpu_3.15", "cpu-time",
		[]stackEntry{{RegularExpression: "^hot$", Percent: &p21}})
	if err := os.WriteFile(filepath.Join(left, "cpu_3.14", "profiles.7.1.info.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(left, "cpu_3.14", "profiles.7.1.internal_metadata.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cmpRun(t, left, right, "", nil); err != nil {
		t.Fatalf("python capture names should load; sidecars ignored: %v", err)
	}
}

func TestCompare_CorruptProfileFails(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 20}})
	if err := os.WriteFile(filepath.Join(left, "cpu_3.14", "profile.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cmpRun(t, left, right, "", nil); err == nil {
		t.Fatal("corrupt profile.json should fail")
	}
}

func TestCompare_DuplicateFamilyFails(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	p20 := int64(20)
	writeCapture(t, left, "cpu_3.14", "cpu-time", []testStack{{"^hot$", 20}})
	writeJSON(t, filepath.Join(left, "cpu_3.14"), "profiles.1.1.json", "cpu_3.14", "cpu-time",
		[]stackEntry{{RegularExpression: "^hot$", Percent: &p20}})
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 20}})
	_, _, err := cmpRun(t, left, right, "", nil)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	first := filepath.Join(left, "cpu_3.14", "profile.json")
	second := filepath.Join(left, "cpu_3.14", "profiles.1.1.json")
	if err == nil || !strings.Contains(msg, "cpu wrote 2 capture JSONs") || !strings.Contains(msg, first) || !strings.Contains(msg, second) || !strings.Contains(msg, "-exclude") {
		t.Fatalf("error must name family, both paths, and -exclude: %v", err)
	}
}

func TestCompare_ZeroFamiliesAfterExcludeFails(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_live_heap_3.14", "heap-live-samples", []testStack{{"^h$", 64}})
	writeCapture(t, right, "python_live_heap_3.15", "heap-live-samples", []testStack{{"^h$", 64}})
	stdout, _, err := cmpRun(t, left, right, "", parseExclude("python_live_heap"))
	if err == nil {
		t.Fatal("0 compared families after excludes should fail")
	}
	if strings.Contains(stdout, "ok:") {
		t.Fatalf("must not print ok: stdout=%q", stdout)
	}
	if !strings.Contains(err.Error(), "no paired families remained after excludes") {
		t.Fatalf("error should say no families remained: %v", err)
	}
}

func TestCompare_MissingSideDir(t *testing.T) {
	right := t.TempDir()
	writeCapture(t, right, "cpu_3.15", "cpu-time", []testStack{{"^hot$", 20}})
	_, _, err := cmpRun(t, filepath.Join(t.TempDir(), "nope"), right, "", nil)
	if err == nil || !strings.Contains(err.Error(), "left: downloads did not produce") || strings.Contains(err.Error(), "lstat") {
		t.Fatalf("missing dir must say downloads without leaking lstat: %v", err)
	}
}
