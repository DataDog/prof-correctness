# Live-heap gate (3.14 baseline)

Pair: `python_live_heap_3.14` / `python_live_heap_3.15`. Asserts persistent live-heap snapshot: `heap-space` and `heap-live-samples` on retained equal-size `bytes` (80/20 count split at `retain_major` / `retain_minor`).

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| heap-space | `retain_major` | 80 | 10 |
| heap-space | `retain_minor` | 20 | 10 |
| heap-live-samples | `retain_major` | 80 | 20 |
| heap-live-samples | `retain_minor` | 20 | 20 |

`allow_first_profile_failure`: first snapshot may predate the full live set.

**Wheel-only** — excluded from prof-correctness `main` CI until live-heap ships on PyPI. With ddtrace `c6197c16` the snapshot is embedded in combined upload profiles (`profiles.*.pprof`); separate `*.heap.pprof` is used when the exporter ships that artifact.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/c6197c1681156e92db986b0b1cdffa08189ff827/install.sh"
TEST_SCENARIOS='python_live_heap_3\.14' go test -v -run TestScenarios -count=1
```
