# python_concurrent_cpu_wall_3.11

Verifies per-thread CPU-vs-wall attribution for threads that are on-CPU and
off-CPU *at the same time*. Unlike `python_spiky_3.11` (single thread that
alternates between bursts and sleeps sequentially), this runs two threads
concurrently for the full duration:

- **busy**: burns CPU in `busy_loop`
- **idle**: loops over short `time.sleep` calls in `idle_wait`

This exercises the sampler reading each thread's own CPU clock
(`clock_gettime(cpu_clock_id)`) independently, and the "wall is guaranteed, CPU
is optional" semantics: a sleeping thread still produces wall-time samples but
~0 CPU.

## Expected behavior

- **cpu-time**: almost all CPU is in `busy_loop` (label `thread name: busy`);
  the idle thread contributes ~0.
- **wall-time**: `busy_loop` and `idle_wait` contribute the *same* amount
  (~33% each, the remaining third being the MainThread blocked in `join`),
  confirming that wall time is attributed equally to the on-CPU and the
  sleeping thread.
