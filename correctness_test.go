package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/prof-correctness/analysis"
	"github.com/DataDog/prof-correctness/reporter"
)

func retrieveCurrentCommand(imageID string) ([]string, error) {
	out, err := exec.Command("docker", "inspect", "--format='{{.Config.Cmd}}'", imageID).Output()
	if err != nil {
		return nil, errors.New("Failed to inspect docker image")
	}
	// split the string into a slice of strings
	cmdSlice := strings.Fields(string(out))
	return cmdSlice, err
}

func runTestApp(t *testing.T, dockerTag string, folder string) string {
	cmdSlice, _ := retrieveCurrentCommand(dockerTag)
	t.Log("Running docker command with output")
	t.Log(strings.Join(cmdSlice, " "))

	tmpdir, err := runTestAppSafe(dockerTag, folder)
	if err != nil {
		t.Fatalf("Error running the test: %v", err)
	}
	t.Log("Docker run output written to", tmpdir)
	return tmpdir
}

// newReporter constructs the metrics recorder for a `testScenarios` run.
// Returns nil and an empty cleanup when the reporter is disabled (default).
// Activated by DD_PROF_CORRECTNESS_REPORT=1; falls back to a no-op recorder
// (and logs a warning) if DD_API_KEY_STAGING is missing.
func newReporter(t *testing.T) (analysis.MetricsSink, func()) {
	if os.Getenv("DD_PROF_CORRECTNESS_REPORT") == "" {
		return nil, func() {}
	}
	apiKey := os.Getenv("DD_API_KEY_STAGING")
	if apiKey == "" {
		t.Logf("DD_PROF_CORRECTNESS_REPORT=1 but DD_API_KEY_STAGING is empty; metrics disabled")
		return nil, func() {}
	}
	site := os.Getenv("DD_SITE")
	if site == "" {
		site = "datad0g.com" // staging by default for this project
	}
	var commonTags []string
	if v := os.Getenv("RUNNER_LABEL"); v != "" {
		commonTags = append(commonTags, "runner_label:"+v)
	}
	if v := os.Getenv("GIT_REPO"); v != "" {
		commonTags = append(commonTags, "git_repo:"+v)
	}
	if v := os.Getenv("GIT_BRANCH"); v != "" {
		commonTags = append(commonTags, "git_branch:"+v)
	}
	rec := reporter.NewDatadogRecorder(apiKey, site, "github-actions", commonTags)
	// Wrap with a mutex so concurrent t.Run goroutines don't race on the
	// recorder's internal buffers (defensive — DatadogRecorder already uses
	// its own locks, but explicit serialisation here documents intent).
	mu := &sync.Mutex{}
	w := &lockedRecorder{rec: rec, mu: mu}
	cleanup := func() {
		if err := rec.Flush(); err != nil {
			t.Logf("reporter flush failed (non-fatal): %v", err)
		}
	}
	return w, cleanup
}

// lockedRecorder adapts *reporter.DatadogRecorder to analysis.MetricsSink and
// also serialises calls so parallel scenarios stay deterministic for tests.
type lockedRecorder struct {
	rec *reporter.DatadogRecorder
	mu  *sync.Mutex
}

func (l *lockedRecorder) RecordAssertion(ev analysis.AssertionEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rec.RecordAssertion(reporter.AssertionEvent{
		Scenario: ev.Scenario, Language: ev.Language, ProfileType: ev.ProfileType,
		Kind: ev.Kind, ErrorPct: ev.ErrorPct, ErrorMargin: ev.ErrorMargin,
		Passed: ev.Passed, AllowFailure: ev.AllowFailure,
	})
}

func (l *lockedRecorder) recordScenarioResult(scenario, language string, failed bool, durationSeconds float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rec.RecordScenarioResult(scenario, language, failed, durationSeconds)
}

func testScenarios(t *testing.T, scenarioRegexp string) {
	t.Logf("Considering only scenarios in %s", scenarioRegexp)
	rootDir := "./scenarios"
	configs, err := findDockerConfigs(rootDir, t, scenarioRegexp)
	if err != nil {
		panic(err)
	}
	if len(configs) == 0 {
		t.Fatalf("No configurations were found with this regexp %s", scenarioRegexp)
	}

	buildBaseImages(t, configs)

	sink, flushOnce := newReporter(t)
	t.Cleanup(flushOnce)

	// Run the tests
	for _, config := range configs {
		t.Run(config.folder, func(t *testing.T) {
			t.Log("Folder:", config.folder)
			t.Log("Json file:", config.jsonFilePath)
			t.Log("Docker file:", config.dockerfilePath)
			tag := buildTestApp(t, config)
			t.Log("Built test app with:", tag)
			pprof_folder := runTestApp(t, tag, config.folder)
			scenarioName := reporter.ScenarioFromFolder(config.folder)
			language := reporter.LanguageFromScenarioFolder(config.folder)
			start := time.Now()
			analysis.AnalyzeResultsWithOpts(t, config.jsonFilePath, pprof_folder, analysis.AnalyzeOptions{
				Sink: sink, Scenario: scenarioName, Language: language,
			})
			durSeconds := time.Since(start).Seconds()
			if lr, ok := sink.(*lockedRecorder); ok {
				lr.recordScenarioResult(scenarioName, language, t.Failed(), durSeconds)
			}
		})
	}
}

var (
	expectedJson = flag.String("expectedJson", "default.json", "Path to the expected JSON file")
	pprofPath    = flag.String("pprofPath", "./", "Path to the directory with the pprof")
)

func TestAnalyze(t *testing.T) {
	flag.Parse()
	analysis.AnalyzeResults(t, *expectedJson, *pprofPath)
}

func TestDDProfScenarios(t *testing.T) {
	testScenarios(t, ".*ddprof.*")
}

func TestPHPScenarios(t *testing.T) {
	testScenarios(t, ".*php.*")
}

func TestAllScenarios(t *testing.T) {
	testScenarios(t, ".*")
}

func TestScenarios(t *testing.T) {
	s := os.Getenv("TEST_SCENARIOS")
	if s != "" {
		testScenarios(t, s)
	} else {
		TestAllScenarios(t)
	}
}

// General Steps
// -- Build test app
// -- Retrieve profilers
// -- Install profiler
// Open question: How do we handle different versions ?
// Knowing that: Profilers can be executables, npm packages, executables
// -- Run test app
// -- Compare results to expected output

// TODO see if we can import tests Felix contributed to Go that checks the block
// profiler bias (basically it was over-reporting some stuff and under-reporting
// other stuff):
// https://cs.opensource.google/go/go/+/master:src/runtime/pprof/pprof_test.go;l=1117-1162
// That file has a bunch of other correctness tests for the profilers, too
