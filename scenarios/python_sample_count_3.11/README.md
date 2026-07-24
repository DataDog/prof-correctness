# python_sample_count_3.11

Verifies that the stack-sampling adaptive sampler produces a reasonable number of
`wall-samples` under its configured overhead target.

## Workload

Uses `asyncio.gather` to run many concurrent coroutines:
- **500 off-cpu tasks**: each calls `asyncio.sleep(0.0001)` 1000 times
- **10 on-cpu tasks**: each calls `math.factorial` 10 times

All coroutines are gathered in a single `asyncio.run` call.

## Expected behavior

- **wall-samples**: the total number of raw stack samples captured over the run is checked
  against a reference value (`value-matching-sum` = 40000) with a wide error margin (20%),
  since the adaptive sampler's interval reacts to CPU usage and host scheduling noise.
  This is a coarse regression check (e.g. catches the sampler firing far too often/rarely),
  not an exact count.
