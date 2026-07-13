#!/usr/bin/env bash
set -euo pipefail

REC=/tmp/rec.jfr

# wall-clock profiling intentionally not enabled: this scenario only exercises
# and asserts on cpu-time/allocation, which is all datadog.yaml's cpu-time and
# allocation mappings need.
java "-agentpath:/app/libjavaProfiler.so=start,cpu=10ms,memory=524288:aL,jfr,file=${REC}" \
    -cp /app CpuAllocWorkload

jbang jafar-tools-latest@btraceio jfr2pprof --config /app/datadog.yaml --output /app/data/profile.pprof "${REC}"
