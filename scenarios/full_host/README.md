# Full Host Profiler Test

This test scenario validates the correctness of the [dd-otel-host-profiler](https://github.com/DataDog/dd-otel-host-profiler) by profiling a simple C application.

## Test Application

The test runs a simple C program (`main.c`) that contains two functions:
- `a()`: Performs 100M iterations of simple arithmetic
- `b()`: Performs 200M iterations of simple arithmetic  

The expected CPU profile should show `b()` taking approximately twice as much CPU time as `a()`.

## Profiler Setup

This test uses the dd-otel-host-profiler, which:
- Runs as a daemon process with elevated privileges
- Profiles all processes on the host using eBPF
- Outputs pprof files to `/app/data/`

## Docker Requirements

The container needs to run with elevated privileges to allow the eBPF-based profiler to function:

```bash
docker run --privileged <image>
```

Or with specific capabilities:
```bash
docker run --cap-add SYS_ADMIN --cap-add SYS_PTRACE <image>
```

## Environment Variables

- `EXECUTION_TIME_SEC`: Duration to run the test application (default: 12 seconds)
The aim is to have 2 profiles, the first one fails due to startup / warmup of full host
- Profiler outputs are written to `/app/data/profiles_*`here