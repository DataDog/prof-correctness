# Alloc gate (3.14 baseline)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Asserts `alloc-space` + `alloc-samples` on two sites with 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`). Workload: `scenarios/python_alloc/main.py`. Memory-only; stack/lock collectors off. `DD_PROFILING_HEAP_SAMPLE_SIZE=65536` and `DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true` are pinned (3.13+ `bytearray` buffers go through `PYMEM_DOMAIN_MEM`). Objects are not retained.

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_1` | 1000000 | 15 |
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_2` | 1000000 | 15 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_1` | 512000000 | 10 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_2` | 1536000000 | 10 |

500k calls × 1024 vs 3072 B yield alloc-space 512/1536 MB. Each `bytearray` is two mallocs (OBJ header + MEM buffer on 3.13+), so alloc-samples is 1e6/site. The 29/39 ratio was mem_domain off (headers only); Dockerfiles pin `DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true`.

Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_alloc_3\.14' go test -v -run TestScenarios
```
