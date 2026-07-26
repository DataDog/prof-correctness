# Python downstream gate (dd-trace-py)

Ten scenarios (**3.14 baseline** / **3.15 candidate**) are the default set when dd-trace-py triggers downstream CI.

| Family | 3.14 | 3.15 | Asserts |
|--------|------|------|---------|
| cpu | `python_cpu_3.14` | `python_cpu_3.15` | cpu-time |
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` | exception-samples |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` | wall-time |
| lock | `python_lock_3.14` | `python_lock_3.15` | lock-acquire/release |
| live-heap | `python_live_heap_3.14` | `python_live_heap_3.15` | heap-space + heap-live-samples (wheel-only) |

Extended families (alloc, asyncio, …) land in follow-up PRs.

## Default regexp

```
python_(cpu|exceptions|async_gen|lock|live_heap)_3\.(14|15)
```

Override via `workflow_dispatch` → `test_scenarios` on [`downstream-python.yml`](../../.github/workflows/downstream-python.yml).

## Wheels & CI

Builds use [`base_images/Dockerfile.python-wheel`](../../base_images/Dockerfile.python-wheel) with `DDTRACE_INSTALL_URL` (S3 wheel from downstream CI).

- `*_3.15` — excluded from `main` CI (`test_scenarios_exclude` in [ci.yml](../../.github/workflows/ci.yml))
- Core `*_3.14` (cpu, exceptions, async_gen, lock) — run on `main` CI (PyPI ddtrace)
- Wheel-only `*_3.14` (live_heap, …) — also in `test_scenarios_exclude` until PyPI ships the feature

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(cpu|exceptions|async_gen|lock|live_heap)_3\.(14|15)' go test -v -run TestScenarios
```

See [README](../../README.md#downstream-from-dd-trace-py).
