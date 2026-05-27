package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// binPath is set by TestMain to point at a freshly-built prof-analyze.
var binPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "prof-analyze-test-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdtemp: %v\n", err)
		os.Exit(2)
	}
	defer os.RemoveAll(tmp)

	binPath = filepath.Join(tmp, "prof-analyze")
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintf(os.Stderr, "go toolchain not found: %v\n", err)
		os.Exit(2)
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Stderr = os.Stderr
	build.Stdout = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build prof-analyze: %v\n", err)
		os.Exit(2)
	}
	os.Exit(m.Run())
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// copyFixturePprof copies testdata/profile.pprof into dir. We can't run the
// CLI directly against testdata/ because captureProfData writes a sibling
// .json file next to every pprof it touches — that would pollute the repo on
// every test run.
func copyFixturePprof(t *testing.T, dir string) {
	t.Helper()
	src, err := os.Open("testdata/profile.pprof")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer src.Close()
	dst, err := os.Create(filepath.Join(dir, "profile.pprof"))
	if err != nil {
		t.Fatalf("create dest: %v", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		t.Fatalf("copy: %v", err)
	}
}

func TestCLI_HappyPath(t *testing.T) {
	dir := t.TempDir()
	copyFixturePprof(t, dir)
	cmd := exec.Command(binPath,
		"-expectedJson", "testdata/expected.json",
		"-pprofPath", dir,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if code := exitCode(err); code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
}

func TestCLI_AssertionFailure(t *testing.T) {
	dir := t.TempDir()
	copyFixturePprof(t, dir)
	cmd := exec.Command(binPath,
		"-expectedJson", "testdata/expected-mismatch.json",
		"-pprofPath", dir,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if code := exitCode(err); code != 1 {
		t.Fatalf("expected exit 1 on mismatch, got %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
}

func TestCLI_MissingFlags(t *testing.T) {
	cmd := exec.Command(binPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if code := exitCode(err); code != 2 {
		t.Fatalf("expected exit 2 on missing flags, got %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
}

func TestCLI_MissingExpectedFile(t *testing.T) {
	dir := t.TempDir()
	copyFixturePprof(t, dir)
	cmd := exec.Command(binPath,
		"-expectedJson", filepath.Join(dir, "does-not-exist.json"),
		"-pprofPath", dir,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if code := exitCode(err); code != 1 {
		t.Fatalf("expected exit 1 on missing expected JSON (Fatalf path), got %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
}
