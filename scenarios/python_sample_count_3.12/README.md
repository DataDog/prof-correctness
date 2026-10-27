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
  (`scale_by_duration`) is checked against a reference value (`value-matching-sum` = 4500)
  with a wide error margin (20%, so 3600 to 5400 per second).
  The ceiling is about 5000 per second: 100 sampling cycles per second (10 ms default interval)
  times 50 reservoir tasks. Nightly CI measured 3900 to 4800 per second with ddtrace 4.15.x
  and 4600 to 4930 with ddtrace 4.16.0. Host scheduling noise still makes this a coarse
  regression check rather than an exact count.
