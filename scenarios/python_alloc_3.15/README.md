# Alloc gate (3.15 candidate)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Same workload and expectations as 3.14 (`scenarios/python_alloc/main.py`, `profile.json`). Memory-only; stack/lock collectors off. `DD_PROFILING_HEAP_SAMPLE_SIZE=65536` and `DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true` are pinned.

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_1` | 1000000 | 15 |
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_2` | 1000000 | 15 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_1` | 512000000 | 10 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_2` | 1536000000 | 10 |

**Wheel-only** — requires `DDTRACE_INSTALL_URL`; excluded from `main` CI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_alloc_3\.15' go test -v -run TestScenarios
```
