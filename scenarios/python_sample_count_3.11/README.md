# python_sample_count_3.11

Verifies that the Python profiler correctly distinguishes CPU-time from wall-time when a process mixes on-CPU work and `time.sleep`, and that the stack-sampling adaptive sampler produces a reasonable number of `wall-samples` under its configured overhead target.

Two threads run concurrently for the same duration:
- **MainThread**: burns CPU with `math.factorial` (on-CPU work)
- **Thread-1**: loops over `time.sleep` (off-CPU/sleeping)

## Expected behavior

- **wall-time**: both threads contribute ~50% each, since both run for the same duration
- **cpu-time**: only `cpu_work` appears (~99%+), since sleeping does not consume CPU
- **wall-samples**: the total number of raw stack samples captured over the run is checked
  against a reference value (`value-matching-sum`) with a wide error margin, since the
  adaptive sampler's interval reacts to CPU usage and host scheduling noise. This is a
  coarse regression check (e.g. catches the sampler firing far too often/rarely), not an
  exact count.

## CPU usage over time

Execution alternates CPU spikes (busy-looping on `math.factorial`, ~100% CPU)
with sleep periods (~0% CPU). Spike/sleep lengths are random, but each half of
the run sums to `EXECUTION_TIME_SEC / 2`, so the mean CPU usage is exactly 50%.

```text
CPU
100% ┤ ┌──┐   ┌─┐  ┌──┐  ┌─┐   ┌──┐  ┌─┐  ┌──┐
     │ │  │   │ │  │  │  │ │   │  │  │ │  │  │
 50% ┼─┼──┼───┼─┼──┼──┼──┼─┼───┼──┼──┼─┼──┼──┼──  mean = 50%
     │ │  │   │ │  │  │  │ │   │  │  │ │  │  │
  0% ┴─┘  └───┘ └──┘  └──┘ └───┘  └──┘ └──┘  └──▶ time
       cpu  sleep cpu sleep ...  (each half sums to 50%)
```
