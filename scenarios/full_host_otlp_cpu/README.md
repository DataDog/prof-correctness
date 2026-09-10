# full_host_otlp_cpu

Full-host **CPU** profiling with the datadog-agent `host-profiler`, exercising
the analyzer's **OTLP** input path end to end. It is the CPU counterpart to the
`full_host` (pprof, standalone dd-otel-host-profiler) scenario.

```
cpu_workload  ->  host-profiler (eBPF CPU)  ->  OTLP export  ->  otlp_dump (.otlp)  ->  analyzer
```

The profiler and the OTLP capture come from the **`prof-fullhost-otlp` base
image** (`base_images/Dockerfile.fullhost-otlp`): the **standalone** full-host
profiler image (`registry.datadoghq.com/ddot-ebpf-dev:devtest-latest` — the
agent `host-profiler` as an OpenTelemetry Collector distribution) with the
generic `otlp_dump` capture sidecar (`tools/otlp-dump`) baked in. So the
scenario is self-contained and thin — just workload + config + start script:
**no agent, no hand-built binary, no `binaries/` prerequisite**. The moving
`devtest-latest` tag is intentional, to surface drift. Run:

```sh
TEST_SCENARIOS="full_host_otlp_cpu" go test -v -run TestScenarios
```

The harness runs any `*full_host*` scenario `--privileged --pid=host` with the
debugfs/tracefs mounts the eBPF profiler needs.

`otlp_dump` (in `tools/otlp-dump`) is a generic, dependency-free OTLP/HTTP
receiver that persists the profiler's export as `.otlp` files; it is reusable
by any OTLP-emitting scenario, not specific to this one.

## What this scenario shows about OTLP vs pprof expectations

This is deliberately a "what does it look like" scenario. Compared with the
pprof `full_host` scenario, the `expected_profile.json` differs in ways that are
inherent to the OTLP host-profiler output, not to the harness:

- **`profile-type` is `samples`** (unit `count`), not `cpu-time` (nanoseconds).
  The value is a sample count.
- **`scale_by_duration: false`** — the per-report sample counts are small, so
  rate-scaling would truncate them toward 0. Assert on raw counts / percent.
- **Frames are mapping basenames** (`cpu_workload`, `libc.so.6`, `linux-vdso.1.so`)
  because symbols aren't uploaded, so one logical stack fragments into several
  entries (`cpu_workload;libc.so.6;libc.so.6;cpu_workload`, `…;linux-vdso.1.so`,
  …). We therefore assert a **regex-contains** on the workload binary
  (`.*cpu_workload.*`) with a `percent` band rather than an exact stack+value.

The assertion is a **load-independent rate**: the workload pins one core, so at
the eBPF sampler frequency (~20 Hz) it accounts for ~20 samples/sec regardless
of what else runs on the host (`value-matching-sum: 20`, `scale_by_duration:
true`). Measured ~100 samples per ~5s steady-state report (0–3% error vs the
30% margin).

**Timing / warm state:** a rate assertion only holds for reports where the
workload is fully on-CPU. `start.sh` therefore runs the workload *longer* than
the profiling window and stops the profiler while the workload is still busy, so
there is no idle "cool-down" tail report. The first report still straddles
profiler start-up (partial), which is why `allow_first_profile_failure` is set.

Config note: the image's Datadog-flavored `otlp_http` exporter requires a
`dd-api-key` header even when pointed locally; the value is unused (otlp_dump
ignores it).

## Open question this raises

The differences above are exactly what the label/expectations design note
(`docs/label-expectations-design.md`) is about: do we converge these onto shared
semantic fields (so one expected file works for pprof and OTLP), or accept some
format-specific expectations? This scenario is a concrete data point for that
discussion.
