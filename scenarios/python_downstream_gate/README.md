# Python downstream gate (dd-trace-py)

Eight prof-correctness scenarios exercise the profiling stack on **3.14
(baseline)** and **3.15 (candidate)** for the same workloads. They are the
default set when dd-trace-py triggers downstream CI on profiling changes.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` |
| lock | `python_lock_3.14` | `python_lock_3.15` |
| mem-domain | `python_mem_domain_3.14` | `python_mem_domain_3.15` |

## Profiler coverage

| Family | Profile type asserted | Collectors / setup |
|--------|----------------------|--------------------|
| exceptions | `exception-samples` + `exception type` | Exception profiler |
| async-gen | `wall-time` | Full profiler via `ddtrace-run`; asyncio async-generator workload |
| lock | `lock-acquire` + `lock-release` + `lock name` | Lock profiler; threaded lock churn |
| mem-domain | `heap-space` | MEM-domain heap profiler; wheel-only until PyPI ships the feature |

Feature-specific pairs (live_heap, …) and extended coverage (cpu, alloc,
asyncio, …) land in follow-up PRs.

## Default downstream regexp

```
python_(exceptions|async_gen|lock|mem_domain)_3\.(14|15)
```

Override via `workflow_dispatch` → `test_scenarios`, or when triggering
[`downstream-python.yml`](../../.github/workflows/downstream-python.yml) manually.

## Base images

Gate scenarios use [`base_images/Dockerfile.python-wheel`](../../base_images/Dockerfile.python-wheel)
with a `PYTHON_IMAGE` build arg (`python:3.14` / `python:3.15.0b1`) and optional
`DDTRACE_INSTALL_URL` for wheel pre-install.

## Wheel install

Every scenario builds against a **dd-trace-py wheel** via `DDTRACE_INSTALL_URL`
(as `downstream-python.yml` does:
`https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh`), pre-installed in
the base image when downstream CI runs.

- **All `*_3.15` folders** — PyPI wheels may not be published for 3.15 yet;
  excluded from prof-correctness `main` CI (see `test_scenarios_exclude` in
  [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)).
- **Core `*_3.14` folders** (exceptions, async_gen, lock) — run on `main` CI
  against PyPI ddtrace; downstream CI still passes `DDTRACE_INSTALL_URL`.
- **Wheel-only 3.14 folders** (mem_domain, live_heap, …) — add to
  `test_scenarios_exclude` in `ci.yml` as they land until the feature ships on PyPI.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(exceptions|async_gen|lock|mem_domain)_3\.(14|15)' go test -v -run TestScenarios
```

## Further reading

- Gate infra: PR stacking from `vlad/gate-infra`
- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
