## CPU stack profiling (3.15 candidate)

Validates that the Datadog Python profiler's **stack collector** reports
`cpu-time` samples with correct call stacks and `thread name` labels under a
CPU-bound workload. This is the **3.15 candidate** half of the `3.14 -> 3.15`
migration pair (see `python_cpu_3.14`).

Reuses the workload from `scenarios/python_cpu`: two tight loops (`a` and `b`)
with a 2:1 relative CPU share. Memory profiling is disabled so the assertion
targets stack samples only.

## Expected profile

- `cpu-time`:
  - `<module>;.*main;.*b` ~= 66% (`MainThread`)
  - `<module>;.*main;.*a` ~= 33% (`MainThread`)

`cpu-time` is best-effort per sample, so `error_margin` is generous.
`scale_by_duration` normalizes across run lengths.

## Notes

This scenario is wheel-only: PyPI ddtrace wheels may not exist for 3.15 yet, so
it is excluded from prof-correctness `main` CI and runs via the dd-trace-py
downstream gate (`DDTRACE_INSTALL_URL`).
