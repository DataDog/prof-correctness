package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".github", "workflows", "ci.yml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func applySed(t *testing.T, program, input string) string {
	t.Helper()
	cmd := exec.Command("sed", program)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("sed %q: %v", program, err)
	}
	return strings.TrimSuffix(string(out), "\n")
}

func TestWorkflow_CaptureArtifactNameContract(t *testing.T) {
	root := repoRoot(t)
	testYML, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "test.yml"))
	if err != nil {
		t.Fatal(err)
	}
	ciYML, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	prefixSed := regexp.MustCompile(`(?m)SAFE_PREFIX=\$\(printf '%s' "\$CAPTURE_ARTIFACT_PREFIX" \| sed '([^']+)'\)`)
	m := prefixSed.FindSubmatch(testYML)
	if m == nil {
		t.Fatal("SAFE_PREFIX sed not found in test.yml")
	}
	got := applySed(t, string(m[1]), "gate-3.14")
	if got != "gate-3.14" {
		t.Fatalf("prefix sanitize must keep '.': gate-3.14 -> %q (sed %s)", got, m[1])
	}

	start := strings.Index(string(ciYML), "name: compare 3.14 vs 3.15")
	end := strings.Index(string(ciYML), "\n  full_host:")
	if start < 0 || end < 0 || end <= start {
		t.Fatal("compare job block not found in ci.yml")
	}
	compare := string(ciYML[start:end])
	if strings.Count(compare, "uses: actions/download-artifact") != 2 {
		t.Fatalf("expected 2 download-artifact steps, got job:\n%s", compare)
	}
	if strings.Count(compare, "pattern: gate-3.14_*") != 1 {
		t.Fatalf("download pattern gate-3.14_* missing:\n%s", compare)
	}
	if strings.Count(compare, "pattern: gate-3.15_*") != 1 {
		t.Fatalf("download pattern gate-3.15_* missing:\n%s", compare)
	}
	if strings.Count(compare, "if-no-artifact-found: fail") != 2 {
		t.Fatalf("both download steps must set if-no-artifact-found: fail:\n%s", compare)
	}
}
