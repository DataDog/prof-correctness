# Python downstream gate (dd-trace-py)

Twenty scenarios (**3.14 baseline** / **3.15 candidate**) are the default set when dd-trace-py triggers downstream CI. The uvloop pair is parked: no cp315 wheel. The gevent pair is parked: greenlet cp315 needs `_PyGC_VisitFrameStack` (3.15b2+), and we stay on `python:3.15.0b1` because the pinned ddtrace cp315 wheel SIGSEGVs on rc1.

| Family | 3.14 | 3.15 | Asserts |
|--------|------|------|---------|
| cpu | `python_cpu_3.14` | `python_cpu_3.15` | cpu-time |
| alloc | `python_alloc_3.14` | `python_alloc_3.15` | alloc-space/samples |
| asyncio | `python_asyncio_3.14` | `python_asyncio_3.15` | wall-time + task name |
| native-cpu | `python_native_cpu_3.14` | `python_native_cpu_3.15` | cpu-time + wall-time (C ext) |
| deep-stack | `python_deep_stack_3.14` | `python_deep_stack_3.15` | cpu-time |
| gil-contention | `python_gil_contention_3.14` | `python_gil_contention_3.15` | cpu-time + wall-time |
| uvloop | `python_uvloop_3.14` | `python_uvloop_3.15` | parked (no 3.15 wheel) |
| gevent | `python_gevent_3.14` | `python_gevent_3.15` | parked (greenlet needs 3.15b2+; ddtrace cp315 SIGSEGVs on rc1) |
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` | exception-samples |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` | wall-time |
| lock | `python_lock_3.14` | `python_lock_3.15` | lock-acquire/release |
| live-heap | `python_live_heap_3.14` | `python_live_heap_3.15` | heap-space + heap-live-samples (wheel-only) |

Feature-specific pairs (mem_domain) land in follow-up PRs.

## Default regexp

```
python_(cpu|alloc|asyncio|native_cpu|deep_stack|gil_contention|exceptions|async_gen|lock|live_heap)_3\.(14|15)
```

Override via `workflow_dispatch` → `test_scenarios` on [`downstream-python.yml`](../../.github/workflows/downstream-python.yml).

## Wheels & CI

Builds use [`base_images/Dockerfile.python-wheel`](../../base_images/Dockerfile.python-wheel) with `DDTRACE_INSTALL_URL` (S3 wheel from downstream CI).

- `python` — old (non-gate) Python scenarios, PyPI ddtrace
- `python_3_14` / `python_3_15` — gate pairs, same pinned S3 wheel (temporary until natives are on main)
- `compare` — diffs 3.14 vs 3.15 captures and fails when percents diverge inside the shared band

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(cpu|alloc|asyncio|native_cpu|deep_stack|gil_contention|exceptions|async_gen|lock|live_heap)_3\.(14|15)' go test -v -run TestScenarios
```

See [README](../../README.md#downstream-from-dd-trace-py).
