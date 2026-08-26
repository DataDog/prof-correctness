# Alloc gate (3.14 baseline)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Asserts `alloc-space` + `alloc-samples` on two sites with 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`). Workload: `scenarios/python_alloc/main.py`. Memory-only; stack/lock collectors off. `DD_PROFILING_HEAP_SAMPLE_SIZE=65536` and `DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true` are pinned (3.13+ `bytearray` buffers go through `PYMEM_DOMAIN_MEM`). Objects are not retained.

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_1` | 1000000 | 15 |
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_2` | 1000000 | 15 |
| alloc-samples matching-sum | both sites | 2000000 | 15 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_1` | 512000000 | 10 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_2` | 1536000000 | 10 |
| alloc-space matching-sum | both sites | 2048000000 | 10 |

500000 calls per site at 1024 vs 3072 bytes. alloc-space is the payload closed form (512/1536 MB); ~56-byte OBJ headers add ~5% and sit inside the margin. alloc-samples is 1e6/site because each `bytearray` is two mallocs (OBJ header + MEM buffer), not 5e5. `scale_by_duration` is false.

With mem_domain off, CI historically observed ~29/39 space (sd=0, n=64) and ~72–105 MB totals — only the headers.

Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_alloc_3\.14' go test -v -run TestScenarios
```
