# Live-heap gate (3.15 candidate)

Pair: `python_live_heap_3.14` / `python_live_heap_3.15`. Same workload and expectations as 3.14; asserts `heap-space` + `heap-live-samples` on the live-heap snapshot (`*.heap.pprof`).

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| heap-space | `retain_major` | 80 | 10 |
| heap-space | `retain_minor` | 20 | 10 |
| heap-live-samples | `retain_major` | 80 | 12 |
| heap-live-samples | `retain_minor` | 20 | 12 |

**Wheel-only** — requires `DDTRACE_INSTALL_URL` at build time; excluded from prof-correctness `main` CI (3.15 + live-heap not on PyPI yet). Runs via downstream gate.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_live_heap_3\.15' go test -v -run TestScenarios
```
