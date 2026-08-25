# Alloc gate (3.14 baseline)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Asserts `alloc-space` + `alloc-samples` on two sites with 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`). Workload: `scenarios/python_alloc/main.py`. Memory-only; stack/lock collectors off.

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_1` | 32 | 5 |
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_2` | 43 | 5 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_1` | 29 | 5 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_2` | 39 | 5 |

Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_alloc_3\.14' go test -v -run TestScenarios
```
