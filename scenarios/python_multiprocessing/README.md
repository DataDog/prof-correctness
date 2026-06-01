# python_multiprocessing

Verifies that the stack sampler survives `fork` and keeps producing correct
profiles in forked child processes. The profiler is started in the parent and
then `NUM_WORKERS` children are forked, so each child inherits the running
native sampler and must restart it through the `pthread_atfork` handler
(re-registering the surviving thread with a fresh clock id and clearing stale
echion state). Only `python_gunicorn_3.11` touches forking today, and it does
so as a web server with `allow_first_profile_failure`, so the pure-compute fork
path is otherwise uncovered.

The parent and every child run the identical CPU-bound `work` loop, so each
process emits a separate pprof file with the same expected shape.

## Expected behavior

- **cpu-time**: in every per-process profile (parent and each forked child),
  `work` accounts for nearly all of that process' CPU. A correct, work-dominated
  profile from each forked child proves the sampler restarted cleanly after the
  fork. The assertion is a percentage, so it holds regardless of how many cores
  the host gives the processes.
