package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// TestFlakiness runs a single scenario N times in parallel to detect flaky tests.
//
// Usage:
//
//	TEST_SCENARIOS="python_basic_3.11" FLAKINESS_RUNS=10 go test -v -timeout 30m -run TestFlakiness
//
// Environment variables:
//   - TEST_SCENARIOS: regex matching the scenario to test (required, should match exactly one)
//   - FLAKINESS_RUNS: number of parallel runs (default: 10)
//   - TEST_RUN_SECS: duration for each run in seconds (default: 60)
func TestFlakiness(t *testing.T) {
	scenarioRegexp := os.Getenv("TEST_SCENARIOS")
	if scenarioRegexp == "" {
		t.Skip("TEST_SCENARIOS not set, skipping flakiness testing")
	}

	numRuns := 10
	if s := os.Getenv("FLAKINESS_RUNS"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			t.Fatalf("Invalid FLAKINESS_RUNS value: %s", s)
		}
		numRuns = n
	}

	t.Logf("Flakiness test: running scenarios matching %q %d times in parallel", scenarioRegexp, numRuns)

	rootDir := "./scenarios"
	configs, err := findDockerConfigs(rootDir, t, scenarioRegexp)
	if err != nil {
		t.Fatalf("Failed to find configs: %v", err)
	}
	if len(configs) == 0 {
		t.Fatalf("No scenarios matched %q", scenarioRegexp)
	}

	if len(configs) > 1 {
		names := make([]string, len(configs))
		for i, c := range configs {
			names[i] = c.folder
		}
		t.Fatalf("TEST_SCENARIOS=%q matched %d scenarios (expected exactly 1): %s",
			scenarioRegexp, len(configs), strings.Join(names, ", "))
	}

	buildBaseImages(t, configs)

	for _, config := range configs {
		t.Run(config.folder, func(t *testing.T) {
			// Build the image once
			tag := buildTestApp(t, config)
			t.Logf("Built image %s, launching %d parallel runs", tag, numRuns)

			// Run all containers in parallel, collect output dirs
			type runResult struct {
				index    int
				pprofDir string
				err      error
			}
			results := make([]runResult, numRuns)
			var wg sync.WaitGroup
			wg.Add(numRuns)
			for i := range numRuns {
				go func(idx int) {
					defer wg.Done()
					defer func() {
						if r := recover(); r != nil {
							results[idx] = runResult{index: idx, err: fmt.Errorf("panic: %v\n%s", r, debug.Stack())}
						}
					}()
					pprofDir, runErr := runTestAppSafe(tag, config.folder)
					results[idx] = runResult{index: idx, pprofDir: pprofDir, err: runErr}
				}(i)
			}
			wg.Wait()

			// Analyze each result via sub-tests, track per-run pass/fail
			passed := 0
			failed := 0
			var failedRuns []string
			for _, res := range results {
				name := fmt.Sprintf("run-%d", res.index)
				ok := t.Run(name, func(t *testing.T) {
					if res.err != nil {
						t.Fatalf("Container failed to run: %v", res.err)
					}
					t.Logf("Analyzing results in %s", res.pprofDir)
					AnalyzeResults(t, config.jsonFilePath, res.pprofDir)
				})
				if ok {
					passed++
				} else {
					failed++
					failedRuns = append(failedRuns, fmt.Sprintf("%s (data: %s)", name, res.pprofDir))
				}
			}

			t.Logf("\n=== Flakiness Summary ===")
			t.Logf("Scenario: %s", config.folder)
			t.Logf("Total runs: %d", numRuns)
			t.Logf("Passed:     %d", passed)
			t.Logf("Failed:     %d", failed)
			t.Logf("Pass rate:  %.1f%%", float64(passed)/float64(numRuns)*100)
			if len(failedRuns) > 0 {
				t.Logf("Failed runs:")
				for _, f := range failedRuns {
					t.Logf("  - %s", f)
				}
			}
		})
	}
}
