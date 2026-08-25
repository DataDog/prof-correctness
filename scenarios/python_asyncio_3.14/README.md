# Asyncio task labels gate (3.14 baseline)

Pair: `python_asyncio_3.14` / `python_asyncio_3.15`. Asserts **wall-time** samples carry asyncio `task name` labels (named and unnamed tasks). Workload: `scenarios/python_asyncio/main.py`.

| Profile type | Stack | Expected % | Margin | Labels |
|--------------|-------|------------|--------|--------|
| wall-time | `my_coroutine` | 30 | 10 | task name: `short_task` |
| wall-time | `my_coroutine` | 61 | 10 | task name: `Task-[0-9]+` |

Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_asyncio_3\.14' go test -v -run TestScenarios
```
