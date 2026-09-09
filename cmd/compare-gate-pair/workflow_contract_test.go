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

func TestWorkflow_CaptureArtifactNameContract(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	testYML, err := os.ReadFile(filepath.Join(root, ".github/workflows/test.yml"))
	if err != nil {
		t.Fatal(err)
	}
	ciYML, err := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`SAFE_PREFIX=.*sed '([^']+)'`).FindSubmatch(testYML)
	if m == nil {
		t.Fatal("SAFE_PREFIX sed not found in test.yml")
	}
	cmd := exec.Command("sed", string(m[1]))
	cmd.Stdin = strings.NewReader("gate-3.14")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != "gate-3.14" {
		t.Fatalf("prefix sanitize must keep '.': %q (sed %s)", got, m[1])
	}
	ci := string(ciYML)
	if !strings.Contains(ci, "pattern: gate-3.14_*") || !strings.Contains(ci, "pattern: gate-3.15_*") {
		t.Fatal("compare downloads must use dotted gate-3.14_* / gate-3.15_*")
	}
	if strings.Count(ci, "if-no-artifact-found: fail") < 2 {
		t.Fatal("both compare downloads must set if-no-artifact-found: fail")
	}
}
