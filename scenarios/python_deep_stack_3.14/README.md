# Deep stack gate (3.14 baseline)

Pair: `python_deep_stack_3.14` / `python_deep_stack_3.15`. Asserts **cpu-time** samples capture a 400-frame recursive stack with omitted-frame marker. Workload: `scenarios/python_deep_stack/main.py`.

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| cpu-time | `<N frames omitted>;recurse;…;burn` | 30000000000 | 5 |

**Source:** recurse(400) then burn for `EXECUTION_TIME_SEC` = 30e9 ns is the only CPU work. Margin is profiler/startup frames.

Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_deep_stack_3\.14' go test -v -run TestScenarios
```
