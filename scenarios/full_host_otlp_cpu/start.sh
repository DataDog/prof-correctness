#!/bin/bash
# full_host_otlp_cpu: run the standalone host-profiler (CPU), capture its OTLP
# export to .otlp files via the otlp_dump sidecar, and run a CPU workload.
#
# Timing matters for a rate assertion: every report we assert on must see the
# workload fully on-CPU. So we run the workload LONGER than the profiling window
# and stop the profiler while the workload is still busy — this avoids an
# idle "cool-down" tail report. The first report still straddles profiler
# start-up (partial), which is why the scenario sets allow_first_profile_failure.
set -u

DATA_DIR=/app/data
mkdir -p "${DATA_DIR}"

PROFILE_SECS="${EXECUTION_TIME_SEC}"          # how long we profile
WORKLOAD_SECS=$(( PROFILE_SECS + 10 ))        # workload outlives the profiler

echo "=== starting otlp_dump (captures OTLP export to .otlp) ==="
SINK_OUT_DIR="${DATA_DIR}" SINK_ADDR="127.0.0.1:4318" /usr/local/bin/otlp_dump &
SINK_PID=$!
sleep 1

echo "=== starting host-profiler (standalone) ==="
# Needs privileges (the harness runs full_host* scenarios with --privileged
# --pid=host and debugfs/tracefs mounts).
/opt/datadog-agent/embedded/bin/host-profiler run --config /app/host-profiler-config.yaml &
PROFILER_PID=$!
sleep 3

echo "=== running cpu_workload for ${WORKLOAD_SECS}s (profiling for ${PROFILE_SECS}s) ==="
DD_SERVICE=cpu_workload_test timeout "${WORKLOAD_SECS}"s /app/cpu_workload &
APP_PID=$!

# Collect reports while the workload is busy, then stop the profiler *before*
# the workload ends so the final flushed report is still fully on-CPU.
sleep "${PROFILE_SECS}"

echo "=== stopping profiler (workload still running) ==="
kill "${PROFILER_PID}" 2>/dev/null || true
wait "${PROFILER_PID}" 2>/dev/null || true

kill "${APP_PID}" 2>/dev/null || true
wait "${APP_PID}" 2>/dev/null || true
sleep 1
kill "${SINK_PID}" 2>/dev/null || true
wait "${SINK_PID}" 2>/dev/null || true

echo "=== output files ==="
ls -la "${DATA_DIR}"
