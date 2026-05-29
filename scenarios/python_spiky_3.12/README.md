# python_spiky_3.12

Verifies that the Python profiler correctly distinguishes CPU-time from wall-time when a process mixes on-CPU work and `time.sleep`.

Two threads run concurrently for the same duration:
- **MainThread**: burns CPU with `math.factorial` (on-CPU work)
- **Thread-1**: loops over `time.sleep` (off-CPU/sleeping)

## Expected behavior

- **wall-time**: both threads contribute ~50% each, since both run for the same duration
- **cpu-time**: only `cpu_work` appears (~90%+), since sleeping does not consume CPU
