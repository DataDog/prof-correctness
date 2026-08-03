# Python downstream gate (dd-trace-py)

Sixteen prof-correctness scenarios exercise the profiling stack on **3.14
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
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` |
| lock | `python_lock_3.14` | `python_lock_3.15` |

## Profiler coverage

| Family | Profile type asserted |
|--------|----------------------|
| cpu (stack) | `cpu-time` + `thread name` |
| alloc (memory) | `alloc-space` + `alloc-samples` |
| asyncio | `wall-time` + `task name` |
| native-cpu | `cpu-time` + `wall-time` (C-extension frames) |
| deep-stack | `cpu-time` (400-frame unwinding) |
| exceptions | `exception-samples` + `exception type` |
| async-gen | `wall-time` |
| lock | `lock-acquire` + `lock-release` + `lock name` |

Feature-specific pairs (mem_domain, live_heap) and cross-cutting integration
scenarios land in follow-up PRs.

## Default downstream regexp

```
python_(cpu|alloc|asyncio|native_cpu|deep_stack|exceptions|async_gen|lock)_3\.(14|15)
```

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(cpu|alloc|asyncio|native_cpu|deep_stack|exceptions|async_gen|lock)_3\.(14|15)' go test -v -run TestScenarios
```

## Further reading

- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
