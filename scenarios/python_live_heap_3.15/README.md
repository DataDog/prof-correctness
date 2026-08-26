# Live-heap gate (3.15 candidate)

Pair: `python_live_heap_3.14` / `python_live_heap_3.15`. Same workload and expectations as 3.14 (`scenarios/python_live_heap/main.py`, `profile.json`).

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| heap-space | `retain_major` | 80 | 10 |
| heap-space | `retain_minor` | 20 | 10 |
| heap-live-samples | `retain_major` | 80 | 20 |
| heap-live-samples | `retain_minor` | 20 | 20 |

Stay on percent: captured 3.14 snapshots do not match absolute 1600/400 or heap-space 32768000 (see 3.14 README). `allow_first_profile_failure`: first snapshot may predate the full live set.

**Wheel-only** — excluded from prof-correctness `main` CI until live-heap ships on PyPI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/c6197c1681156e92db986b0b1cdffa08189ff827/install.sh"
TEST_SCENARIOS='python_live_heap_3\.15' go test -v -run TestScenarios -count=1
```

**3.15 blocked** on wheel `c6197c16`: S3 index has cp315 musllinux wheels only; manylinux base image cannot install.
