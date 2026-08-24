# Alloc gate (3.14 baseline)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Asserts `alloc-space` + `alloc-samples` on two sites with 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`). Memory-only; stack/lock collectors off.

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| alloc-samples | `allocate_memory_1` | 33 | 10 |
| alloc-samples | `allocate_memory_2` | 44 | 10 |
| alloc-space | `allocate_memory_1` | 30 | 10 |
| alloc-space | `allocate_memory_2` | 40 | 10 |

Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_alloc_3\.14' go test -v -run TestScenarios
```
