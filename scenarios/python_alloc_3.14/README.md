# Alloc gate (3.14 baseline)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Asserts `alloc-space` + `alloc-samples` on two sites with 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`). Workload: `scenarios/python_alloc/main.py`. Memory-only; stack/lock collectors off.

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_1` | 1e6 | 15 |
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_2` | 1e6 | 15 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_1` | 512 MB | 10 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_2` | 1536 MB | 10 |

Runs on `main` CI (PyPI ddtrace).

Equal call counts and 1:3 byte sizes.

```sh
TEST_SCENARIOS='python_alloc_3\.14' go test -v -run TestScenarios
```
