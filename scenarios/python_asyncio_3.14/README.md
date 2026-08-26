# Asyncio task labels gate (3.14 baseline)

Pair: `python_asyncio_3.14` / `python_asyncio_3.15`. Asserts **wall-time** samples carry asyncio `task name` labels. Workload: `scenarios/python_asyncio/main.py`.

| Profile type | Stack | Expected | Margin | Labels |
|--------------|-------|----------|--------|--------|
| wall-time | `my_coroutine` | 2500000000 | 10 | task name: `short_task` |
| wall-time | `my_coroutine` | 5000000000 | 10 | task name: `long_task` |

**Source:** concurrent sleeps `EXECUTION_TIME_SEC/2` = 2.5e9 ns and `EXECUTION_TIME_SEC` = 5e9 ns. Matching sum 7.5e9 ns.

```sh
TEST_SCENARIOS='python_asyncio_3\.14' go test -v -run TestScenarios
```
