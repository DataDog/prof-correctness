## Allocation profiling (3.14 baseline)

Validates that the Datadog Python profiler's **memory collector** reports
`alloc-space` and `alloc-samples` with correct call stacks on Python 3.14
(`Target.run` / `Target.allocate_memory_*` frames). This is the **3.14 baseline**
half of the `3.14 -> 3.15` migration pair (see `python_alloc_3.15`).

Reuses the workload from `scenarios/python_basic_memory_3.11`: two allocation
sites with a 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`).
Stack and lock collectors are disabled to isolate allocation samples.

## Expected profile

On 3.14, alloc stacks use `Target.method` frames (not bare `run;allocate_memory_*`).
Alloc profiles use numeric thread IDs, so assertions omit `thread name` labels.
Sampling proportions follow `python_basic_memory_3.11` (stable `alloc-samples`
and `alloc-space` ratios).

## Notes

The 3.14 scenario runs on prof-correctness `main` CI (PyPI ddtrace). The 3.15
candidate runs only via the dd-trace-py downstream gate (`DDTRACE_INSTALL_URL`).
