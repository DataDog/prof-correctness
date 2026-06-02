package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

// mkScenarios creates the given directory names under a fresh temp dir and
// returns its path. Also creates a "not_a_dir" plain file to verify it's
// filtered out.
func mkScenarios(t *testing.T, names []string) string {
	t.Helper()
	root := t.TempDir()
	for _, n := range names {
		if err := os.Mkdir(filepath.Join(root, n), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", n, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "not_a_dir"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write not_a_dir: %v", err)
	}
	return root
}

func TestRun_ChunksAlphabetically(t *testing.T) {
	root := mkScenarios(t, []string{
		// Intentionally unsorted to verify run() sorts before chunking.
		"python_basic_3.11",
		"python_asyncio_3.11",
		"python_basic_3.10",
		"python_cpu",
		"python_deep_stack_3.11",
		"node_heap", // must be filtered out by the pattern
	})

	got, err := run("python.*", root, 3)
	if err != nil {
		t.Fatal(err)
	}

	want := []matrixEntry{
		{
			Shard: "1/2",
			Regex: "(^|/)(python_asyncio_3.11|python_basic_3.10|python_basic_3.11)$",
			Names: "python_asyncio_3.11, python_basic_3.10, python_basic_3.11",
		},
		{
			Shard: "2/2",
			Regex: "(^|/)(python_cpu|python_deep_stack_3.11)$",
			Names: "python_cpu, python_deep_stack_3.11",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}
}

func TestRun_SingleChunkWhenSmall(t *testing.T) {
	root := mkScenarios(t, []string{"dotnet_wall", "dotnet_alloc"})

	got, err := run("dotnet.*", root, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d: %#v", len(got), got)
	}
	if got[0].Shard != "1/1" {
		t.Errorf("shard: got %q want 1/1", got[0].Shard)
	}
	if got[0].Regex != "(^|/)(dotnet_alloc|dotnet_wall)$" {
		t.Errorf("regex: got %q", got[0].Regex)
	}
	if got[0].Names != "dotnet_alloc, dotnet_wall" {
		t.Errorf("names: got %q", got[0].Names)
	}
}

func TestRun_AnchoringRejectsSubstringMatches(t *testing.T) {
	// Pattern "python_cpu" must NOT pick up "python_cpu_sleep_sync_3.12".
	root := mkScenarios(t, []string{
		"python_cpu",
		"python_cpu_sleep_sync_3.12",
	})

	got, err := run("python_cpu", root, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Regex != "(^|/)(python_cpu)$" {
		t.Fatalf("expected only python_cpu, got %#v", got)
	}
}

func TestRun_ExactChunkSizeBoundary(t *testing.T) {
	// 6 scenarios with chunk size 3 → exactly 2 chunks of 3 (no remainder).
	root := mkScenarios(t, []string{
		"a", "b", "c", "d", "e", "f",
	})
	got, err := run(".*", root, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(got))
	}
	if got[0].Shard != "1/2" || got[1].Shard != "2/2" {
		t.Errorf("shard labels: %q %q", got[0].Shard, got[1].Shard)
	}
}

func TestRun_NoMatchIsError(t *testing.T) {
	root := mkScenarios(t, []string{"python_cpu"})

	_, err := run("ruby.*", root, 3)
	if err == nil {
		t.Fatal("expected error when no scenarios match")
	}
}

func TestRun_InvalidPatternIsError(t *testing.T) {
	root := mkScenarios(t, []string{"python_cpu"})

	if _, err := run("[invalid", root, 3); err == nil {
		t.Fatal("expected error on invalid regex")
	}
}

// TestRun_RegexMatchesIntendedDirAndOnlyThat verifies the produced regex
// behaves correctly against the Go test runner's matching surface, which is
// filepath.Dir of each Dockerfile (e.g. "scenarios/python_cpu") — substring
// matches must NOT occur.
func TestRun_RegexMatchesIntendedDirAndOnlyThat(t *testing.T) {
	root := mkScenarios(t, []string{
		"python_cpu",
		"python_cpu_sleep_sync_3.12",
		"python_basic_3.11",
	})
	got, err := run("python.*", root, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(got))
	}

	re := regexp.MustCompile(got[0].Regex)
	cases := []struct {
		path string
		want bool
	}{
		{"scenarios/python_cpu", true},
		{"scenarios/python_cpu_sleep_sync_3.12", true},
		{"scenarios/python_basic_3.11", true},
		// substring within another name must NOT match
		{"scenarios/python_cpu_something_else", false},
		// completely unrelated
		{"scenarios/ruby_basic", false},
		// bare name (no "scenarios/" prefix) must also match — the alternation
		// anchor is (^|/) so a leading match works too.
		{"python_cpu", true},
	}
	for _, c := range cases {
		if got := re.MatchString(c.path); got != c.want {
			t.Errorf("MatchString(%q) = %v, want %v (regex=%s)", c.path, got, c.want, re)
		}
	}
}

// TestRun_OutputIsValidJSON encodes the output the same way main() does and
// verifies it round-trips to the expected shape.
func TestRun_OutputIsValidJSON(t *testing.T) {
	root := mkScenarios(t, []string{"a", "b", "c", "d"})
	got, err := run(".*", root, 3)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded []matrixEntry
	if err := json.Unmarshal(buf, &decoded); err != nil {
		t.Fatalf("unmarshal: %v\njson: %s", err, buf)
	}
	if !reflect.DeepEqual(decoded, got) {
		t.Fatalf("round-trip mismatch:\n got=%#v\nwant=%#v", decoded, got)
	}
}
