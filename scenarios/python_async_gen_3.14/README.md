# Async generator gate (3.14 baseline)

Pair: `python_async_gen_3.14` / `python_async_gen_3.15`. Workload: `scenarios/python_async_gen/main.py` (named `gen_task` runs `async_gen_work` for `EXECUTION_TIME_SEC`).

| Profile type | Stack | Expected | Margin | Labels |
|--------------|-------|----------|--------|--------|
| wall-time | `.*async_gen_work.*` | 8000000000 | 20 | task name: `gen_task` |

**Source:** only user coroutine; `EXECUTION_TIME_SEC` = 8e9 ns of wall-time. Margin is scheduler/`sleep(0)`.

```sh
TEST_SCENARIOS='python_async_gen_3\.14' go test -v -run TestScenarios
```
