# Python downstream gate (dd-trace-py)

Twenty-two prof-correctness scenarios exercise the profiling stack on **3.14
(baseline)** and **3.15 (candidate)** for the same workloads. They are the
default set when dd-trace-py triggers downstream CI on profiling changes.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| cpu (stack) | `python_cpu_3.14` | `python_cpu_3.15` |
| alloc (memory) | `python_alloc_3.14` | `python_alloc_3.15` |
| asyncio (task labels) | `python_asyncio_3.14` | `python_asyncio_3.15` |
| native-cpu | `python_native_cpu_3.14` | `python_native_cpu_3.15` |
| deep-stack | `python_deep_stack_3.14` | `python_deep_stack_3.15` |
| gil-contention | `python_gil_contention_3.14` | `python_gil_contention_3.15` |
| uvloop | `python_uvloop_3.14` | `python_uvloop_3.15` |
| gevent | `python_gevent_3.14` | `python_gevent_3.15` |
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` |
| lock | `python_lock_3.14` | `python_lock_3.15` |

## Default downstream regexp

```
python_(cpu|alloc|asyncio|native_cpu|deep_stack|gil_contention|uvloop|gevent|exceptions|async_gen|lock)_3\.(14|15)
```

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(cpu|alloc|asyncio|native_cpu|deep_stack|gil_contention|uvloop|gevent|exceptions|async_gen|lock)_3\.(14|15)' go test -v -run TestScenarios
```

Feature-specific pairs (`mem_domain`, `live_heap`) are separate PRs stacked on
`vlad/gate-infra`.

## Further reading

- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
