# python_gil_contention_3.11

Verifies the wall-vs-CPU relationship across many threads that are
all runnable at once: wall time scales with the number of threads, but CPU time
is bounded by the GIL to roughly one core.

The workload runs 8 CPU-bound `spin` threads for the full duration. 8 is below
the default `max_threads` cap, so every thread is sampled each cycle (no
reservoir sampling, unlike `python_many_threads_3.11`).

## Expected behavior

- **cpu-time**: the combined CPU attributed to `spin` is ~1 core
  (one core-second per second), because the GIL serializes Python bytecode
  execution.
- **wall-time**: the combined wall time attributed to `spin` is ~8 cores
  (one core-second per second per thread), because all 8 threads are alive and
  runnable for the whole run.
