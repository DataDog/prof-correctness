## Allocation profiling (3.15 candidate)

Validates that the Datadog Python profiler's **memory collector** reports
`alloc-space` and `alloc-samples` with correct call stacks and `thread name`
labels. This is the **3.15 candidate** half of the `3.14 -> 3.15` migration pair
(see `python_alloc_3.14`).

Reuses the workload from `scenarios/python_basic_memory_3.11`: two allocation
sites with a 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`).
Stack and lock collectors are disabled to isolate allocation samples.

## Expected profile

Assertions match `python_basic_memory_3.11` (stable sampling proportions for
`alloc-samples` and `alloc-space`).

## Notes

This scenario is wheel-only: PyPI ddtrace wheels may not exist for 3.15 yet, so
it is excluded from prof-correctness `main` CI and runs via the dd-trace-py
downstream gate (`DDTRACE_INSTALL_URL`).
