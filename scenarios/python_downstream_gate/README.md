# Python downstream gate (dd-trace-py)

Ten prof-correctness scenarios exercise the profiling stack on **3.14
(baseline)** and **3.15 (candidate)** for the same workloads. They are the
default set when dd-trace-py triggers downstream CI on profiling changes.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| cpu (stack) | `python_cpu_3.14` | `python_cpu_3.15` |
| alloc (memory) | `python_alloc_3.14` | `python_alloc_3.15` |
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` |
| lock | `python_lock_3.14` | `python_lock_3.15` |

## Profiler coverage

| Family | Profile type asserted | Collectors / setup |
|--------|----------------------|--------------------|
| cpu (stack) | `cpu-time` + `thread name` | Stack via `ddtrace-run`; memory off; CPU-bound loops |
| alloc (memory) | `alloc-space` + `alloc-samples` | Memory profiler; stack/lock off |
| exceptions | `exception-samples` + `exception type` | Exception profiler |
| async-gen | `wall-time` | Full profiler via `ddtrace-run`; asyncio async-generator workload |
| lock | `lock-acquire` + `lock-release` + `lock name` | Lock profiler; threaded lock churn |

Feature-specific pairs (mem_domain, live_heap) and extended coverage land in
follow-up PRs.

## Default downstream regexp

```
python_(cpu|alloc|exceptions|async_gen|lock)_3\.(14|15)
```

Override via `workflow_dispatch` → `test_scenarios`, or when triggering
[`downstream-python.yml`](../../.github/workflows/downstream-python.yml) manually.

## Wheel install

- **All `*_3.15` folders** use `DDTRACE_INSTALL_URL` (excluded from `main` CI).
- **3.14 folders** run on prof-correctness `main` CI against PyPI ddtrace.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(cpu|alloc|exceptions|async_gen|lock)_3\.(14|15)' go test -v -run TestScenarios
```

## Further reading

- Gate infra: PR stacking from `vlad/gate-infra`
- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
