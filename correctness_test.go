package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
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

	// Run the tests
	for _, config := range configs {
		t.Run(config.folder, func(t *testing.T) {
			t.Log("Folder:", config.folder)
			t.Log("Json file:", config.jsonFilePath)
			t.Log("Docker file:", config.dockerfilePath)
			tag := buildTestApp(t, config)
			t.Log("Built test app with:", tag)
			pprof_folder := runTestApp(t, tag, config.folder)
			AnalyzeResults(t, config.jsonFilePath, pprof_folder)
		})
	}
}

var (
	expectedJson = flag.String("expectedJson", "default.json", "Path to the expected JSON file")
	pprofPath    = flag.String("pprofPath", "./", "Path to the directory with the pprof")
)

func TestAnalyze(t *testing.T) {
	flag.Parse()
	AnalyzeResults(t, *expectedJson, *pprofPath)
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
