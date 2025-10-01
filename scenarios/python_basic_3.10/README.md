# Python Basic 3.10 Profiling Test

This test validates that the Datadog Python profiler correctly instruments multi-threaded applications.

## Test Application
- Creates two threads that each run `target()` function for 2 seconds
- MainThread: executes `target(2)` directly  
- Worker Thread: spawns `Thread-1 (target)` that executes `target(2)`

## Expected Profile
The profiler should capture wall-time for both threads:
- `^<module>;target$` from MainThread: ~1.5 seconds
- `^_bootstrap;thread_bootstrap_inner;_bootstrap_inner;run;target$` from Thread-1: ~1.5 seconds

## TODO
- Investigate why wall-time shows ~1.5 seconds instead of expected 2 seconds
