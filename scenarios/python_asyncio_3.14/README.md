# Asyncio task labels gate (3.14 baseline)

Pair: `python_asyncio_3.14` / `python_asyncio_3.15`. Asserts **wall-time** samples carry asyncio `task name` labels (named and unnamed tasks). Reuses `scenarios/python_asyncio_3.11` workload.

```sh
TEST_SCENARIOS='python_asyncio_3\.14' go test -v -run TestScenarios
```
