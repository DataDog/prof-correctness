package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestFullHostOTLPStart(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required")
	}
	source, err := os.ReadFile("scenarios/full_host_otlp_cpu/start.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name              string
		exitRole          string
		exitAfterWorkload bool
		exitStatus        string
		shutdownStatus    string
		wantError         string
	}{
		{name: "success"},
		{name: "sink exits", exitRole: "sink", wantError: "otlp_dump exited"},
		{name: "profiler exits cleanly", exitRole: "profiler", wantError: "host-profiler exited"},
		{name: "profiler crashes", exitRole: "profiler", exitStatus: "2", wantError: "host-profiler exited"},
		{name: "profiler exits during profiling", exitRole: "profiler", exitAfterWorkload: true, wantError: "host-profiler exited"},
		{name: "sink exits during profiling", exitRole: "sink", exitAfterWorkload: true, wantError: "otlp_dump exited"},
		{name: "workload exits", exitRole: "workload", wantError: "cpu_workload exited"},
		{name: "shutdown fails", shutdownStatus: "2", wantError: "host-profiler failed during shutdown: 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			writeExecutable := func(name, body string) string {
				t.Helper()
				path := filepath.Join(dir, name)
				if err := os.WriteFile(path, []byte("#!"+bash+"\n"+body), 0o755); err != nil {
					t.Fatal(err)
				}
				return path
			}
			stub := `
set -eu
printf '%s\n' "$$" > "${OTLP_TEST_DIR}/${role}.pid"
trap 'touch "${OTLP_TEST_DIR}/${role}.stopped"' EXIT
stop() {
    if [[ "${role}" == profiler ]]; then
        read -r app_pid < "${OTLP_TEST_DIR}/workload.pid"
        kill -0 "${app_pid}" || exit 3
        touch "${OTLP_TEST_DIR}/workload-alive-at-shutdown"
        exit "${OTLP_TEST_SHUTDOWN_STATUS:-0}"
    fi
    exit 0
}
trap stop TERM
if [[ "${role}" == workload ]]; then
    printf '%s\n' "${EXECUTION_TIME_SEC}" > "${OTLP_TEST_DIR}/workload-duration"
fi
if [[ "${role}" == "${OTLP_TEST_EXIT_ROLE}" ]]; then
    if [[ "${OTLP_TEST_EXIT_AFTER_WORKLOAD}" == true ]]; then
        while [[ ! -f "${OTLP_TEST_DIR}/workload.pid" ]]; do sleep 0.05; done
    fi
    exit "${OTLP_TEST_EXIT_STATUS:-0}"
fi
while true; do sleep 0.05; done
`
			sink := writeExecutable("sink", "role=sink\n"+stub)
			profiler := writeExecutable("profiler", "role=profiler\n"+stub)
			workload := writeExecutable("workload", "role=workload\n"+stub)
			writeExecutable("timeout", "shift\nexec \"$@\"\n")
			script := strings.NewReplacer(
				"/app/data", filepath.Join(dir, "data"),
				"/usr/local/bin/otlp_dump", sink,
				"/opt/datadog-agent/embedded/bin/host-profiler", profiler,
				"/app/cpu_workload", workload,
			).Replace(string(source))
			path := writeExecutable("start.sh", script)
			t.Cleanup(func() {
				for _, role := range []string{"sink", "profiler", "workload"} {
					if _, err := os.Stat(filepath.Join(dir, role+".stopped")); err == nil {
						continue
					}
					data, err := os.ReadFile(filepath.Join(dir, role+".pid"))
					if err != nil {
						continue
					}
					if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
						if process, err := os.FindProcess(pid); err == nil {
							_ = process.Kill()
						}
					}
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, bash, path)
			cmd.Env = append(os.Environ(),
				"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"EXECUTION_TIME_SEC=2",
				"OTLP_TEST_DIR="+dir,
				"OTLP_TEST_EXIT_ROLE="+tc.exitRole,
				"OTLP_TEST_EXIT_AFTER_WORKLOAD="+strconv.FormatBool(tc.exitAfterWorkload),
			)
			for key, value := range map[string]string{
				"OTLP_TEST_EXIT_STATUS":     tc.exitStatus,
				"OTLP_TEST_SHUTDOWN_STATUS": tc.shutdownStatus,
			} {
				if value != "" {
					cmd.Env = append(cmd.Env, key+"="+value)
				}
			}
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("script timed out: %s", out)
			}
			if tc.wantError != "" {
				if err == nil || !strings.Contains(string(out), tc.wantError) {
					t.Fatalf("error = %v, want %q\n%s", err, tc.wantError, out)
				}
			} else {
				if err != nil {
					t.Fatalf("script failed: %v\n%s", err, out)
				}
				duration, err := os.ReadFile(filepath.Join(dir, "workload-duration"))
				if err != nil || strings.TrimSpace(string(duration)) != "12" {
					t.Fatalf("workload duration = %q, error = %v; want 12", duration, err)
				}
				if _, err := os.Stat(filepath.Join(dir, "workload-alive-at-shutdown")); err != nil {
					t.Fatal("workload was not running when the profiler stopped")
				}
			}
			for _, role := range []string{"sink", "profiler", "workload"} {
				if _, err := os.Stat(filepath.Join(dir, role+".pid")); err == nil {
					if _, err := os.Stat(filepath.Join(dir, role+".stopped")); err != nil {
						t.Errorf("%s was not cleaned up", role)
					}
				}
			}
		})
	}
}
