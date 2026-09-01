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
	raw, err := json.Marshal(captureFile{
		TestName: testName,
		Stacks:   []typedStacks{{ProfileType: profileType, StackContent: content}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func writeAsserted(t *testing.T, scenariosDir, family, profileType, regex string) {
	t.Helper()
	dir := filepath.Join(scenariosDir, family+"_3.14")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	raw, err := json.Marshal(captureFile{
		TestName: family,
		Stacks: []typedStacks{{
			ProfileType:  profileType,
			StackContent: []stackEntry{{RegularExpression: regex}},
		}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "expected_profile.json"), raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestCompare_DivergenceFails(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-ts-a", "python_cpu", "cpu-time", []testStack{{"^hot$", 18}})
	writeCapture(t, right, "python_cpu_3.15-ts-b", "python_cpu", "cpu-time", []testStack{{"^hot$", 24}})
	var stdout, stderr bytes.Buffer
	err := run(runConfig{leftDir: left, rightDir: right, maxPP: 5, stdout: &stdout, stderr: &stderr})
	if err == nil {
		t.Fatal("expected divergence to fail")
	}
	if !strings.Contains(stderr.String(), "|18-24|=6 > 5") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCompare_Unmatched(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-ts-a", "python_cpu", "cpu-time", []testStack{
		{"^hot$", 20}, {"^only14$", 8},
	})
	writeCapture(t, right, "python_cpu_3.15-ts-b", "python_cpu", "cpu-time", []testStack{{"^hot$", 20}})
	var stdout, stderr bytes.Buffer
	err := run(runConfig{leftDir: left, rightDir: right, maxPP: 5, stdout: &stdout, stderr: &stderr})
	if err == nil || !strings.Contains(stderr.String(), "only14") || !strings.Contains(stderr.String(), "unmatched") {
		t.Fatalf("expected unmatched ≥5%% fail: err=%v stderr=%q", err, stderr.String())
	}

	left2, right2 := t.TempDir(), t.TempDir()
	writeCapture(t, left2, "python_cpu_3.14-ts-a", "python_cpu", "cpu-time", []testStack{
		{"^hot$", 20}, {"^noise$", 3},
	})
	writeCapture(t, right2, "python_cpu_3.15-ts-b", "python_cpu", "cpu-time", []testStack{{"^hot$", 20}})
	stdout.Reset()
	stderr.Reset()
	if err := run(runConfig{leftDir: left2, rightDir: right2, maxPP: 5, stdout: &stdout, stderr: &stderr}); err != nil {
		t.Fatalf("tail <5%% should be ignored: %v\nstderr=%s", err, stderr.String())
	}
}

func TestCompare_FactorialAlias(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_native_cpu_3.14-ts-a", "python_native_cpu", "cpu-time", []testStack{
		{`^<module>;main;factorial_work;math\.factorial$`, 21},
	})
	writeCapture(t, right, "python_native_cpu_3.15-ts-b", "python_native_cpu", "cpu-time", []testStack{
		{`^<module>;main;factorial_work;math\.integer\.factorial$`, 20},
	})
	var stdout, stderr bytes.Buffer
	if err := run(runConfig{leftDir: left, rightDir: right, maxPP: 5, stdout: &stdout, stderr: &stderr}); err != nil {
		t.Fatalf("alias should pair: %v\nstderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "3.14=21 3.15=20") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestCompare_AssertedOnly(t *testing.T) {
	left, right, scenarios := t.TempDir(), t.TempDir(), t.TempDir()
	writeAsserted(t, scenarios, "python_cpu", "cpu-time", ".*hot")
	writeCapture(t, left, "python_cpu_3.14-ts-a", "python_cpu", "cpu-time", []testStack{
		{"^<module>;hot$", 20}, {"^CodeProvenance$", 40}, {".*sleep$", 30},
	})
	writeCapture(t, right, "python_cpu_3.15-ts-b", "python_cpu", "cpu-time", []testStack{
		{"^<module>;hot$", 22}, {"^pathlib$", 50},
	})
	var stdout, stderr bytes.Buffer
	err := run(runConfig{
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

func TestCompare_Exclude(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_exceptions_3.14-ts-a", "python_exceptions", "cpu-time", []testStack{
		{`^<module>;main;handle_value_error$`, 41},
	})
	writeCapture(t, right, "python_exceptions_3.15-ts-b", "python_exceptions", "cpu-time", []testStack{
		{`^<module>;main;handle_value_error$`, 47},
	})
	writeCapture(t, left, "python_live_heap_3.14-ts-a", "python_live_heap", "heap-live-samples", []testStack{
		{`^<module>;main;Target\.run;Target\.retain_major$`, 64},
	})
	writeCapture(t, right, "python_live_heap_3.15-ts-b", "python_live_heap", "heap-live-samples", []testStack{
		{`^<module>;main;Target\.run;Target\.retain_major$`, 79},
	})
	writeCapture(t, left, "python_cpu_3.14-ts-a", "python_cpu", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "python_cpu_3.15-ts-b", "python_cpu", "cpu-time", []testStack{{"^hot$", 21}})

	var stdout, stderr bytes.Buffer
	err := run(runConfig{
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

func TestCompare_MissingFamily(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	writeCapture(t, left, "python_cpu_3.14-ts-a", "python_cpu", "cpu-time", []testStack{{"^hot$", 20}})
	writeCapture(t, right, "python_alloc_3.15-ts-b", "python_alloc", "alloc-space", []testStack{{"^alloc$", 40}})
	var stdout, stderr bytes.Buffer
	err := run(runConfig{leftDir: left, rightDir: right, maxPP: 5, stdout: &stdout, stderr: &stderr})
	if err == nil || (!strings.Contains(stderr.String(), "missing on right") && !strings.Contains(stderr.String(), "missing on left")) {
		t.Fatalf("expected missing family: err=%v stderr=%q", err, stderr.String())
	}
}

func TestFamilyFromName(t *testing.T) {
	if got := familyFromName("python_cpu_3.14-20260101-aaa"); got != "python_cpu" {
		t.Fatalf("got %q", got)
	}
	if got := familyFromName("python_live_heap_3.15-bbb"); got != "python_live_heap" {
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
	var stdout, stderr bytes.Buffer
	err := run(runConfig{leftDir: t.TempDir(), rightDir: t.TempDir(), maxPP: 5, stdout: &stdout, stderr: &stderr})
	if err == nil || !strings.Contains(err.Error(), "no capture JSON") {
		t.Fatalf("err=%v", err)
	}
}
