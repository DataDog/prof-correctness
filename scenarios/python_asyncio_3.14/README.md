# Asyncio task labels gate (3.14 baseline)

Pair: `python_asyncio_3.14` / `python_asyncio_3.15`. Asserts **wall-time** samples carry asyncio `task name` labels. Workload: `scenarios/python_asyncio/main.py`.

| Profile type | Stack | Expected % | Margin | Labels |
|--------------|-------|------------|--------|--------|
| wall-time | `my_coroutine` | 33 | 10 | task name: `short_task` |
| wall-time | `my_coroutine` | 67 | 10 | task name: `Task-[0-9]+` |

**Source:** concurrent sleeps `EXECUTION_TIME_SEC/2` and `EXECUTION_TIME_SEC` → 1:2 → 33/67.

```sh
TEST_SCENARIOS='python_asyncio_3\.14' go test -v -run TestScenarios
```
