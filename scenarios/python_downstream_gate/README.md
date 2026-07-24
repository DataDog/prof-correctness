# Python downstream gate (dd-trace-py)

Eight prof-correctness scenarios exercise the profiling stack on **3.14
(baseline)** and **3.15 (candidate)** for the same workloads. They are the
default set when dd-trace-py triggers downstream CI on profiling changes.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| mem-domain | `python_mem_domain_3.14` | `python_mem_domain_3.15` |
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` |
| lock | `python_lock_3.14` | `python_lock_3.15` |

## Profiler coverage

Each family asserts one **profile type** end-to-end (collector + export + pprof
shape). Together they exercise four collectors on the 3.14 → 3.15 path — **not**
the full Python profiling matrix.

| Family | Profile type asserted | Collectors / setup | Not exercised in this scenario |
|--------|----------------------|--------------------|--------------------------------|
| mem-domain | `heap-space` | Heap + **MEM domain** (`DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED`); stack/lock off | CPU stack, wall-time, lock, exceptions, heap-live |
| exceptions | `exception-samples` | **Exception** profiler (`DD_PROFILING_EXCEPTION_ENABLED`) | Other collectors at default/disabled |
| async-gen | `wall-time` | Full profiler via `ddtrace-run` (`DD_PROFILING_ENABLED`); asyncio async-generator workload | Explicit CPU-only assertion; lock/heap/exception not targeted |
| lock | `lock-acquire` | **Lock** profiler (`DD_PROFILING_LOCK_ENABLED`); threaded lock churn | Stack, heap, exceptions |

**Not covered by this gate** (other prof-correctness scenarios or dd-trace-py riot
tests): CPU/stack as the primary signal, heap-live samples, PyTorch, gevent/uwsgi
integration shapes, Flask/FastAPI HTTP workloads, native alloc, deep stack, GIL
contention, etc. Expand the downstream regexp to `python.*` when ready for
broader E2E coverage.

## Default downstream regexp

GitLab and GitHub triggers in dd-trace-py both pass the same regexp (not the
downstream workflow default of `python.*`):

```
python_(mem_domain|exceptions|async_gen|lock)_3\.(14|15)
```

Override via `workflow_dispatch` → `test_scenarios`, or when triggering
`downstream-python.yml` manually.

## Wheel install (3.15)

3.15 scenarios require a **dd-trace-py wheel** at image build time:

- Set `DDTRACE_INSTALL_URL` when running tests (as `downstream-python.yml` does:
  `https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh`).
- Do **not** `pip install ddtrace` from PyPI in the scenario Dockerfile — 3.15
  wheels may not be published yet.

3.15 folders are excluded from prof-correctness `main` CI until 3.15 wheels are
generally available; they run via the dd-trace-py downstream gate.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(mem_domain|exceptions|async_gen|lock)_3\.(14|15)' go test -v -run TestScenarios
```

## Further reading

- Full migration playbook (unit / E2E / staging A/B):
  `experimental/teams/profiling-python/ddtrace-upgrade/ab_staging_experiments_design.md`
- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
