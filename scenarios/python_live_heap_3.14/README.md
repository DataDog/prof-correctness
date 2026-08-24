# Live-heap gate (3.14 baseline)

Pair: `python_live_heap_3.14` / `python_live_heap_3.15`. Asserts persistent live-heap snapshot (`*.heap.pprof`): `heap-space` and `heap-live-samples` on retained equal-size `bytes` (80/20 at `retain_major` / `retain_minor`).

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| heap-space | `retain_major` | 80 | 10 |
| heap-space | `retain_minor` | 20 | 10 |
| heap-live-samples | `retain_major` | 80 | 12 |
| heap-live-samples | `retain_minor` | 20 | 12 |

`allow_first_profile_failure`: first snapshot may predate the full live set.

**Wheel-only** — excluded from prof-correctness `main` CI until live-heap ships on PyPI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_live_heap_3\.14' go test -v -run TestScenarios
```
