# python_many_threads_3.11

Verifies that the stack sampler's reservoir sampling produces correct wall-time
totals when the number of live threads exceeds the per-cycle sampling cap.

The workload spawns 20 identical CPU-bound `worker` threads that all run for the
full duration. `_DD_PROFILING_STACK_MAX_THREADS` forces the sampler to sample
only a subset of the 21 threads each cycle (Algorithm R reservoir sampling),
scaling each sampled thread's wall time by `n_total / sample_count`.

## Expected behavior

- **wall-time**: the combined wall time attributed to `worker` is ~20 cores
  (one core-second per second per worker). Despite subsampling, the
  inverse-probability weighting must reconstruct this total. This is the only
  scenario that crosses the `max_threads` cap and exercises reservoir sampling.
