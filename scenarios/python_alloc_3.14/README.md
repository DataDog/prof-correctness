# Alloc gate (3.14 baseline)

Pair: `python_alloc_3.14` / `python_alloc_3.15`. Asserts `alloc-space` + `alloc-samples` on two sites with 1:3 byte ratio (`allocate_memory_1` / `allocate_memory_2`). Workload: `scenarios/python_alloc/main.py`. Memory-only; stack/lock collectors off.

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_1` | 32 | 5 |
| alloc-samples | `<module>;.*Target.run;Target.allocate_memory_2` | 43 | 5 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_1` | 29 | 5 |
| alloc-space | `<module>;.*Target.run;Target.allocate_memory_2` | 39 | 5 |

Runs on `main` CI (PyPI ddtrace).

**Attribution gap (do not retune onto 29/39):** `main.py` documents a 1:3 byte ratio so alloc-space closed form is 25/75. Observed 3.14 alloc-space is an invariant **29:39 (1:1.345, sd=0 across 64 CI runs)** and the two sites together are only 68% of alloc-space despite being ~99.6% of allocated bytes. `profile_3.15.json` already asserts 25/75. Investigate before treating either number as correct.

```sh
TEST_SCENARIOS='python_alloc_3\.14' go test -v -run TestScenarios
```
