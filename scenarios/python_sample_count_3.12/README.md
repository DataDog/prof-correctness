# python_sample_count_3.12

Verifies that the stack sampler produces a reasonable number of `wall-samples`
with task reservoir sampling enabled.

## Workload

Uses `asyncio.gather` to run many concurrent coroutines:
- **500 off-cpu tasks**: each calls `asyncio.sleep(0.0001)` 1000 times
- **10 on-cpu tasks**: each calls `math.factorial` 10 times

All coroutines are gathered in a single `asyncio.run` call.

## Expected behavior

- **Task reservoir**: by default, the profiler samples at most 50 leaf tasks per cycle.
  This scenario intentionally exercises that default behavior.
- **wall-samples**: the number of raw stack samples per second of profile
  (`scale_by_duration`) is checked against the theoretical rate (`value-matching-sum` = 5000):
  100 sampling cycles per second (10 ms default interval) times 50 reservoir tasks.
  Real runs land below it, for example when the sampler thread wakes late on a busy host.
  The error margin is wide (30%, so 3500 to 6500 per second): CI measured 3667 to 4931 per
  second with ddtrace 4.15.x, 4.16.0 and dd-trace-py `main`. This is a coarse check that
  catches a sampler running far too often or too rarely.
