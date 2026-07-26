# Python downstream gate (dd-trace-py)

Two prof-correctness scenarios exercise **persistent live-heap** profiling on
**3.14 (baseline)** and **3.15 (candidate)**. Both dirs are **wheel-only**
(persistent live-heap is not in a published PyPI release yet).

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| live-heap | `python_live_heap_3.14` | `python_live_heap_3.15` |

## Default downstream regexp

```
python_live_heap_3\.(14|15)
```

## CI exclude

Both folders require `DDTRACE_INSTALL_URL`. `python_live_heap_3.14` is also
listed in [`ci.yml`](../../.github/workflows/ci.yml) `test_scenarios_exclude`
until the feature ships in a PyPI release.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_live_heap_3\.(14|15)' go test -v -run TestScenarios
```
