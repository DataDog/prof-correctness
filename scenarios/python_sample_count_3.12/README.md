# python_sample_count_3.12

Verifies that the stack sampler produces a reasonable number of `wall-samples`
with task reservoir sampling enabled.

## Workload

Uses `asyncio.gather` to run many concurrent coroutines:
- **500 off-cpu tasks**: each calls `asyncio.sleep(0.0001)` 1000 times
- **10 on-cpu tasks**: each calls `math.factorial` 10 times

All coroutines are gathered in a single `asyncio.run` call.

## Expected behavior

- **Task reservoir**: at most 50 leaf tasks are sampled per cycle. The explicit cap keeps
  the expectation independent of changes to the profiler's default configuration.
- **wall-samples**: the total number of raw stack samples captured over the run is checked
  against a reference value (`value-matching-sum` = 4000) with a wide error margin (20%).
  Reservoir sampling reduces the prior count by approximately `50 / 500`; host scheduling
  noise still makes this a coarse regression check rather than an exact count.
