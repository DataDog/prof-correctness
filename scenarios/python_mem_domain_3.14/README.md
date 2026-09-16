# MEM-domain gate (3.14 baseline)

Pair: `python_mem_domain_3.14` / `python_mem_domain_3.15`. Asserts `heap-space` on a retained 16 MiB `bytearray` at `allocate_mem_domain_buffer` (PYMEM_DOMAIN_MEM on 3.13+). Workload: `scenarios/python_mem_domain/main.py`. Memory-only; stack/lock collectors off. Feature flag not pinned (default-on).

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| heap-space | `allocate_mem_domain_buffer` | 16777216 | 20 |

**Wheel-only** — MEM-domain default-on is not in PyPI 4.14.1. Requires `DDTRACE_INSTALL_URL`; excluded from prof-correctness `main` CI PyPI job.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_mem_domain_3\.14' go test -v -run TestScenarios
```
