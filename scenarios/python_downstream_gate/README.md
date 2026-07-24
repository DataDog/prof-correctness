# Python downstream gate (dd-trace-py)

Ten prof-correctness scenarios exercise the profiling stack on **3.14
(baseline)** and **3.15 (candidate)** for the same workloads. They are the
default set when dd-trace-py triggers downstream CI on profiling changes.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| mem-domain | `python_mem_domain_3.14` | `python_mem_domain_3.15` |
| exceptions | `python_exceptions_3.14` | `python_exceptions_3.15` |
| async-gen | `python_async_gen_3.14` | `python_async_gen_3.15` |
| lock | `python_lock_3.14` | `python_lock_3.15` |
| live-heap | `python_live_heap_3.14` | `python_live_heap_3.15` |

## Profiler coverage

Each family asserts one **profile type** end-to-end (collector + export + pprof
shape). Together they exercise five collectors on the 3.14 → 3.15 path — **not**
the full Python profiling matrix.

| Family | Profile type asserted | Collectors / setup | Not exercised in this scenario |
|--------|----------------------|--------------------|--------------------------------|
| mem-domain | `heap-space` | Heap + **MEM domain** (`DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED`); stack/lock off | CPU stack, wall-time, lock, exceptions, heap-live |
| exceptions | `exception-samples` | **Exception** profiler (`DD_PROFILING_EXCEPTION_ENABLED`) | Other collectors at default/disabled |
| async-gen | `wall-time` | Full profiler via `ddtrace-run` (`DD_PROFILING_ENABLED`); asyncio async-generator workload | Explicit CPU-only assertion; lock/heap/exception not targeted |
| lock | `lock-acquire` | **Lock** profiler (`DD_PROFILING_LOCK_ENABLED`); threaded lock churn | Stack, heap, exceptions |
| live-heap | `heap-space` + `heap-live-samples` | **Persistent live-heap** snapshot (`DD_PROFILING_HEAP_ENABLED`); OBJ-domain `bytes`; stack/lock off | CPU stack, wall-time, lock, exceptions, alloc-space |

**Not covered by this gate** (other prof-correctness scenarios or dd-trace-py riot
tests): CPU/stack as the primary signal, PyTorch, gevent/uwsgi integration
shapes, Flask/FastAPI HTTP workloads, native alloc, deep stack, GIL contention,
etc. Expand the downstream regexp to `python.*` when ready for broader E2E
coverage.

## Default downstream regexp

GitLab and GitHub triggers in dd-trace-py both pass the same regexp (not the
downstream workflow default of `python.*`):

```
python_(mem_domain|exceptions|async_gen|lock|live_heap)_3\.(14|15)
```

Override via `workflow_dispatch` → `test_scenarios`, or when triggering
`downstream-python.yml` manually.

## Wheel install

Every scenario builds against a **dd-trace-py wheel** via `DDTRACE_INSTALL_URL`
(as `downstream-python.yml` does:
`https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh`), pre-installed in
the base image. Two cases need it unconditionally:

- **All 3.15 folders** — PyPI wheels may not be published for 3.15 yet.
- **Both live-heap folders (3.14 and 3.15)** — the persistent live-heap profile
  is recent and not yet in a published release.

These folders are excluded from prof-correctness `main` CI (which uses PyPI
ddtrace) until the relevant wheels are generally available; they run via the
dd-trace-py downstream gate. The other 3.14 folders also run on `main` CI.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_(mem_domain|exceptions|async_gen|lock|live_heap)_3\.(14|15)' go test -v -run TestScenarios
```

## Future work

This gate tests the **migration delta** (3.14 → 3.15). A separate axis worth
adding is **floor coverage**: run the same workloads on the oldest supported
interpreter (currently 3.9; 3.10 once 3.15 drops 3.9) to catch regressions for
users who jump directly from an old version to the newest — e.g. 3.9 → 3.15.
Note prof-correctness only ships base images for 3.10–3.15 today, so a true 3.9
floor scenario would need a new base image. This is intentionally out of scope
for the 3.14/3.15 delta gate.

## Further reading

- Full migration playbook (unit / E2E / staging A/B):
  `experimental/teams/profiling-python/ddtrace-upgrade/ab_staging_experiments_design.md`
- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
