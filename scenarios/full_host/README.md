# Full Host Profiler Test

This test scenario validates the correctness of the [dd-otel-host-profiler](https://github.com/DataDog/dd-otel-host-profiler) by profiling a simple C application.

## Test Application

The test runs a simple C program (`main.c`) that contains two functions:
- `a()`: Performs 100M iterations of simple arithmetic
- `b()`: Performs 200M iterations of simple arithmetic  

The expected CPU profile will not contain symbols.
We could check that the unwinding is happening correctly (separate locations).
Though for now the test frameowrk does not support it.

## Profiler Setup

This test uses the dd-otel-host-profiler, which:
- Runs as a daemon process with elevated privileges
- Profiles all processes on the host using eBPF
- With split by service, we filter out the data that is relevant to the test (refer to json file for `pprof-regex`)
- We ignore the first profile (as full host has a warm up time of ~20ms)

```
test prog        <------------------>       (12 sec after sleep of 2 seconds)
upload period <---------><--------->        (5 sec x2)
```

## Environment Variables

- `EXECUTION_TIME_SEC`: Duration to run the test application (default: 12 seconds)
The aim is to have 2 profiles, the first one fails due to startup / warmup of full host
- Profiler outputs are written to `/app/data/profiles_*`here

## Other ideas

We could have tests that start within the upload period which would cover the startup time.

```
test prog        <---------------->       (9 sec)
upload period <------------------------>  (10 sec)
```
