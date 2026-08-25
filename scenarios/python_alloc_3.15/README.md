# Alloc gate (3.15 candidate)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Same workload (`scenarios/python_alloc/main.py`); expectations from `python_alloc/profile_3.15.json` (3.15 uses bare `run;allocate_memory_*` stacks and `thread name` labels).

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| alloc-samples | `run;allocate_memory_1` | 35 | 5 |
| alloc-samples | `run;allocate_memory_2` | 45 | 5 |
| alloc-samples | `__init__;grow_list` | 19 | 5 |
| alloc-space | `run;allocate_memory_1` | 25 | 5 |
| alloc-space | `run;allocate_memory_2` | 75 | 5 |
| alloc-space | `grow_list` | 3 | 3 |

**Wheel-only** — requires `DDTRACE_INSTALL_URL`; excluded from `main` CI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_alloc_3\.15' go test -v -run TestScenarios
```
