package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var RUN_SECS = uint(60)
var DURATION_SET = false
var NETWORK_HOST = false

func init() {
	s := os.Getenv("TEST_RUN_SECS")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			panic("Invalid value for env var TEST_RUN_SECS")
		}
		RUN_SECS = uint(i)
		DURATION_SET = true
	}
	network_host, ok := os.LookupEnv("NETWORK_HOST")
	if ok && network_host != "OFF" {
		NETWORK_HOST = true
	}
}

type DockerTestConfig struct {
	folder         string
	jsonFilePath   string
	dockerfilePath string
}

func findDockerConfigs(rootDir string, t *testing.T, scenarioRegexp string) ([]DockerTestConfig, error) {
	var configs []DockerTestConfig
	var folder, jsonFilePath, dockerfilePath string
	testPathRegexp := regexp.MustCompile(scenarioRegexp)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !testPathRegexp.MatchString(filepath.Dir(path)) {
			return nil
		}

		// check if the file is a JSON file or a Dockerfile
		if filepath.Base(path) == "expected_profile.json" {
			jsonFilePath = path
		} else if filepath.Base(path) == "Dockerfile" {
			dockerfilePath = path
		} else {
			// skip files that are not JSON or Dockerfiles
			return nil
		}
		// if we have both a JSON file and a Dockerfile, create a Config instance
		if jsonFilePath != "" && dockerfilePath != "" {
			if filepath.Dir(jsonFilePath) != filepath.Dir(dockerfilePath) {
				t.Errorf("miss matching file structure in %s", filepath.Dir(jsonFilePath))
				return nil
			}
			folder = filepath.Dir(jsonFilePath)
			configs = append(configs, DockerTestConfig{folder, jsonFilePath, dockerfilePath})

			jsonFilePath = ""
			dockerfilePath = ""
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func retrieveCurrentCommand(imageID string) ([]string, error) {
	out, err := exec.Command("docker", "inspect", "--format='{{.Config.Cmd}}'", imageID).Output()
	if err != nil {
		return nil, errors.New("Failed to inspect docker image")
	}
	// split the string into a slice of strings
	cmdSlice := strings.Fields(string(out))
	return cmdSlice, err
}

// returns the tag for built docker app
func buildTestApp(t *testing.T, config DockerTestConfig) string {
	// we could use the docker client, though that makes it harder to do command lines manually
	now_time := time.Now()
	// Following arg helps forces to rerun steps after the arg (allows reinstallation of recent profiler) --build-arg CACHE_DATE=$(date +%Y-%m-%d_%H:%M:%S)
	cmd := exec.Command("docker", "build", "-f", config.dockerfilePath, "--build-arg", now_time.Format("2006-01-02_15:04:05"), "-t", "test-app", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", err)
		t.Fatalf("Error building %s - %s", config.folder, out)
	}
	return string("test-app")
}

// docker run -v ${PWD}/data:/app/data:rw -e EXECUTION_TIME=60 -u $(id -u ${USER}):$(id -g ${USER}) --security-opt seccomp=unconfined test-app:latest

func runTestApp(t *testing.T, dockerTag string, folder string) string {
	currentPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Outputs are written to /app/data
	profilePath := currentPath + "/data"
	tmpdir, err := os.MkdirTemp(profilePath, filepath.Base(folder)+"-*")
	if err != nil {
		t.Fatalf("Failed to make tmp dir: %v", err)
	}
	mountOption := tmpdir + ":/app/data:rw"
	// ensure we run with the same user (so we can read the profiles)
	userID := os.Getuid()
	groupID := os.Getgid()
	userOption := fmt.Sprintf("%d:%d", userID, groupID)

	cmdSlice, _ := retrieveCurrentCommand(dockerTag)
	t.Log("Running docker command with output", tmpdir)
	t.Log(strings.Join(cmdSlice, " "))
	args := []string{"run", "-v", mountOption, "-u", userOption, "--security-opt", "seccomp=unconfined"}
	if DURATION_SET {
		args = append(args, "-e", "EXECUTION_TIME="+fmt.Sprint(RUN_SECS))
	}
	if NETWORK_HOST {
		args = append(args, "--network=host")
	}
	args = append(args, "-e", "DD_SERVICE=prof-correctness-"+folder)
	args = append(args, "test-app:latest")
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error running the test %s - %s", strings.Join(cmd.Args, " "), out)
	}
	// Dump the combined output to a file as it might contain useful information
	// such as tracebacks or error messages from the profiler that don't
	// necessarily cause the test to fail.
	err = ioutil.WriteFile(tmpdir+"/output.txt", out, 0644)
	if err != nil {
		t.Fatalf("Failed to write output to file: %v", err)
	}
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
	for _, config := range configs {
		t.Log("Folder:", config.folder)
		t.Log("Json file:", config.jsonFilePath)
		t.Log("Docker file:", config.dockerfilePath)
		tag := buildTestApp(t, config)
		t.Log("Built test app with:", tag)
		pprof_folder := runTestApp(t, tag, config.folder)
		AnalyzeResults(t, config.jsonFilePath, pprof_folder)
	}
}

var (
	expectedJson = flag.String("expectedJson", "default.json", "Path to the expected JSON file")
	pprofPath = flag.String("pprofPath", "./", "Path to the directory with the pprof")
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
