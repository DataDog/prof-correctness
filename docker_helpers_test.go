package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	RUN_SECS     = uint(60)
	DURATION_SET = false
	NETWORK_HOST = false
)

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

var (
	baseImageArgPattern     = regexp.MustCompile(`ARG BASE_IMAGE="(.+?)"`)
	baseImageArgDeclPattern = regexp.MustCompile(`^\s*ARG\s+BASE_IMAGE\s*$`)
)

// Find an inline BASE_IMAGE default in the scenario Dockerfile, if any.
func extractBaseImage(dockerfilePath string) (string, error) {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 10 {
		line := scanner.Text()
		matches := baseImageArgPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1], nil
		}
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func baseImageFromFolder(folder string) (string, error) {
	name := filepath.Base(folder)
	switch {
	case strings.HasSuffix(name, "_3.14"):
		return "prof-python-3.14", nil
	case strings.HasSuffix(name, "_3.15"):
		return "prof-python-3.15", nil
	default:
		return "", fmt.Errorf("no BASE_IMAGE in Dockerfile and folder %q has no known _3.14/_3.15 suffix", folder)
	}
}

func resolveBaseImage(config DockerTestConfig) (string, error) {
	baseImage, err := extractBaseImage(config.dockerfilePath)
	if err != nil {
		return "", err
	}
	if baseImage != "" {
		return baseImage, nil
	}
	if !dockerfileDeclaresBaseImageArg(config.dockerfilePath) {
		return "", nil
	}
	return baseImageFromFolder(config.folder)
}

func dockerfileDeclaresBaseImageArg(dockerfilePath string) bool {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 10 {
		line := scanner.Text()
		if baseImageArgPattern.MatchString(line) || baseImageArgDeclPattern.MatchString(line) {
			return true
		}
		lineCount++
	}
	return false
}

// buildBaseImages collects and builds all base images referenced by the given configs.
func buildBaseImages(t *testing.T, configs []DockerTestConfig) {
	baseImageNames := map[string]bool{}
	for _, config := range configs {
		baseImage, err := resolveBaseImage(config)
		if err != nil {
			t.Fatalf("Error resolving base image for %s: %v", config.folder, err)
		}
		if baseImage != "" {
			baseImageNames[baseImage] = true
		}
	}
	for baseImageName := range baseImageNames {
		t.Log("Building base image:", baseImageName)
		buildBaseImage("./", baseImageName, t)
	}
}

func buildBaseImage(rootDir string, baseImageName string, t *testing.T) {
	baseImageDir := filepath.Join(rootDir, "base_images")

	pythonWheelImages := map[string]string{
		"prof-python-3.14": "python:3.14",
		"prof-python-3.15": "python:3.15.0rc1",
	}

	var dockerfilePath string
	args := []string{"build", "-t", baseImageName}

	if pythonImage, ok := pythonWheelImages[baseImageName]; ok {
		dockerfilePath = filepath.Join(baseImageDir, "Dockerfile.python-wheel")
		args = append(args, "-f", dockerfilePath, "--build-arg", "PYTHON_IMAGE="+pythonImage)
	} else {
		dockerfileName := "Dockerfile." + strings.TrimPrefix(baseImageName, "prof-")
		dockerfilePath = filepath.Join(baseImageDir, dockerfileName)
		args = append(args, "-f", dockerfilePath)
	}

	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Fatalf("Required base Dockerfile %s not found!", dockerfilePath)
		return
	}

	if u := os.Getenv("DDTRACE_INSTALL_URL"); u != "" {
		args = append(args, "--build-arg", "DDTRACE_INSTALL_URL="+u)
	}
	args = append(args, rootDir)
	buildCmd := exec.Command("docker", args...)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Error building base image %s: %v", baseImageName, err)
	}
	t.Logf("Built base image with tag: %s", baseImageName)
}

// returns the tag for built docker app
func buildTestApp(t *testing.T, config DockerTestConfig) string {
	resolvedDockerfile, err := filepath.EvalSymlinks(config.dockerfilePath)
	if err != nil {
		t.Fatalf("Error resolving Dockerfile symlink for %s: %v", config.folder, err)
	}
	config.dockerfilePath = resolvedDockerfile

	// we could use the docker client, though that makes it harder to do command lines manually
	now_time := time.Now()
	// Following arg helps forces to rerun steps after the arg (allows reinstallation of recent profiler) --build-arg CACHE_DATE=$(date +%Y-%m-%d_%H:%M:%S)
	args := []string{"build", "-f", config.dockerfilePath, "--build-arg", "CACHE_DATE=" + now_time.Format("2006-01-02_15:04:05"), "-t", "test-app"}
	if baseImage, err := resolveBaseImage(config); err != nil {
		t.Fatalf("Error resolving base image for %s: %v", config.folder, err)
	} else if baseImage != "" && dockerfileDeclaresBaseImageArg(config.dockerfilePath) {
		args = append(args, "--build-arg", "BASE_IMAGE="+baseImage)
	}
	if u := os.Getenv("DDTRACE_INSTALL_URL"); u != "" {
		args = append(args, "--build-arg", "DDTRACE_INSTALL_URL="+u)
	}
	args = append(args, ".")
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", err)
		t.Fatalf("Error building %s - %s", config.folder, out)
	}
	return string("test-app")
}

// runTestAppSafe runs a docker container for the given scenario and returns the
// output directory. It returns an error instead of calling t.Fatalf, making it
// safe to call from goroutines.
func runTestAppSafe(dockerTag string, folder string) (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	profilePath := currentPath + "/data"
	timestamp := time.Now().Format("20060102-150405")
	tmpdir, err := os.MkdirTemp(profilePath, filepath.Base(folder)+"-"+timestamp+"-*")
	if err != nil {
		return "", fmt.Errorf("mkdtemp: %w", err)
	}
	mountOption := tmpdir + ":/app/data:rw"
	userID := os.Getuid()
	groupID := os.Getgid()
	userOption := fmt.Sprintf("%d:%d", userID, groupID)

	var args []string
	if strings.Contains(folder, "full_host") {
		args = []string{"run", "--rm", "-v", mountOption, "--pid=host", "--privileged", "--security-opt", "seccomp=unconfined"}
		args = append(args, "--cap-add=SYS_ADMIN", "--cap-add=SYS_PTRACE", "--cap-add=SYS_RESOURCE")
		args = append(args, "-v", "/sys/kernel/debug:/sys/kernel/debug:ro")
		args = append(args, "-v", "/sys/kernel/tracing:/sys/kernel/tracing:ro")
	} else {
		args = []string{"run", "--rm", "-v", mountOption, "-u", userOption, "--security-opt", "seccomp=unconfined"}
	}

	if DURATION_SET {
		args = append(args, "-e", "EXECUTION_TIME_SEC="+fmt.Sprint(RUN_SECS))
	}
	if NETWORK_HOST {
		args = append(args, "--network=host")
	}
	if !strings.Contains(folder, "full_host") {
		args = append(args, "-e", "DD_SERVICE=prof-correctness-"+strings.Split(folder, "/")[1])
	}
	args = append(args, dockerTag+":latest")

	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.WriteFile(tmpdir+"/output.txt", out, 0644) // best-effort
		return tmpdir, fmt.Errorf("docker run failed: %w\noutput: %s", err, out)
	}
	if writeErr := os.WriteFile(tmpdir+"/output.txt", out, 0644); writeErr != nil {
		return tmpdir, fmt.Errorf("write output: %w", writeErr)
	}
	return tmpdir, nil
}

func TestResolveBaseImageFromFolder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		folder         string
		dockerfilePath string
		want           string
	}{
		{
			folder:         "scenarios/python_async_gen_3.14",
			dockerfilePath: "scenarios/python_async_gen/Dockerfile",
			want:           "prof-python-3.14",
		},
		{
			folder:         "scenarios/python_async_gen_3.15",
			dockerfilePath: "scenarios/python_async_gen/Dockerfile",
			want:           "prof-python-3.15",
		},
		{
			folder:         "scenarios/python_lock_3.14",
			dockerfilePath: "scenarios/python_lock/Dockerfile",
			want:           "prof-python-3.14",
		},
		{
			folder:         "scenarios/python_lock_3.15",
			dockerfilePath: "scenarios/python_lock/Dockerfile",
			want:           "prof-python-3.15",
		},
		{
			folder:         "scenarios/python_exceptions_3.14",
			dockerfilePath: "scenarios/python_exceptions/Dockerfile",
			want:           "prof-python-3.14",
		},
		{
			folder:         "scenarios/python_exceptions_3.15",
			dockerfilePath: "scenarios/python_exceptions/Dockerfile",
			want:           "prof-python-3.15",
		},
		{
			folder:         "scenarios/python_live_heap_3.14",
			dockerfilePath: "scenarios/python_live_heap/Dockerfile",
			want:           "prof-python-3.14",
		},
		{
			folder:         "scenarios/python_live_heap_3.15",
			dockerfilePath: "scenarios/python_live_heap/Dockerfile",
			want:           "prof-python-3.15",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.folder, func(t *testing.T) {
			t.Parallel()
			config := DockerTestConfig{
				folder:         tc.folder,
				dockerfilePath: tc.dockerfilePath,
			}
			baseImage, err := resolveBaseImage(config)
			if err != nil {
				t.Fatalf("resolveBaseImage: %v", err)
			}
			if baseImage != tc.want {
				t.Fatalf("got base image %q, want %q", baseImage, tc.want)
			}
		})
	}
}
