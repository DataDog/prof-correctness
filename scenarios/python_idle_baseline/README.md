# python_idle_baseline

Verifies the adaptive-sampling **baseline floor**: the minimum CPU budget that
keeps the stack sampler running even when the application is idle.
`python_spiky_3.11` enables adaptive sampling but is ~50% busy, so its baseline
floor never binds; this scenario isolates it.

The workload is a single thread that does nothing but `time.sleep` for the whole
run (`idle_wait`). The adaptive sampler is configured with a short
`p_stable_window` (so the idle period dominates the stable-CPU estimate within
~1s and the startup CPU ages out) and a 100ms max interval (so without a floor
the sampler backs off to ~10Hz). The baseline floor is then the only thing that
can keep the sample rate up.

This assertion relies on the `min_value` bound and the `wall-samples` count
profile type: summed `wall-time`/`cpu-time` totals telescope (each sample
carries the delta since the previous one) and therefore can't reveal the
sampling rate, but the `wall-samples` count can.

## Expected behavior

- **wall-samples**: the idle thread accumulates at least 1000 wall-time samples.
  Measured locally, the baseline floor produces ~9-10k samples over 10s, versus
  ~130 with the floor disabled, so the 1000-sample minimum fails if the floor
  stops working.
