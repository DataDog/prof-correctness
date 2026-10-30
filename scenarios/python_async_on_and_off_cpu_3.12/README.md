# python_async_on_and_off_cpu_3.12

_**Note** This workload currently does not behave as expected._
_We should only see a minority of CPU time attributed to `off_cpu_task`,_
_in practice we see 99% of CPU time attributed to it. This is a fundamental_
_limitation of tick-based profiling for spiky workloads, and something_
_we plan to fix in the future with timer-based profiling._

Verifies that the Python profiler correctly distinguishes CPU-time from wall-time
for asyncio tasks that mix on-CPU and off-CPU work.

All tasks run concurrently via `asyncio.gather` on a single thread:
- **50 x off_cpu_task**: each loops 1000 times over `asyncio.sleep(0.001)`
- **1 x on_cpu_task**: loops 100 times over `math.factorial(3500)` (CPU-bound)

## Expected behavior

- **wall-time**: dominated by `off_cpu_task` (~99%), since the 50 sleeping tasks
  accumulate far more wall-clock time than the single CPU-bound task
- **cpu-time**: `on_cpu_task` gets 10% +/- 6%, verifying that CPU-time is attributed to
  asyncio task names. Without the profiler, `on_cpu_task` uses about 17 ms of the 155-190 ms
  of CPU in `asyncio.run()`, 9-11% (CPython 3.12, `time.thread_time()`). The rest is event loop
  overhead from the sleeping tasks. The profiler samples every 10 ms and gives each sample the
  CPU time since the previous one, so the 17 ms block can be off by about one interval, or
  6 points.
