# gevent gate (3.14 baseline)

Pair: `python_gevent_3.14` / `python_gevent_3.15`. Workload: `scenarios/python_gevent/main.py` (Hub + 10 `threading.Thread` greenlets running `target()`).

Per-greenlet wall-time is not a stable property: a sample-interval deficit lands on a different `Greenlet-N` each run. Assert the **sum** of all worker greenlets.

| Profile type | Stack | Expected | Margin | Labels |
|--------------|-------|----------|--------|--------|
| wall-time | `^Hub\.run;gevent\.libev\.corecext\.loop\.run$` | 2000000000 | 12 | task name: `Hub` |
| wall-time | `.*target.*` | 10000000000 | 12 | task name `values_regex`: `^Greenlet-[0-9]+$` |

**Source:** Hub = `EXECUTION_TIME_SEC` = 2e9 ns. Workers = `10 × EXECUTION_TIME_SEC/2` = 10e9 ns (`scale_by_duration: false`).

```sh
TEST_SCENARIOS='python_gevent_3\.14' go test -v -run TestScenarios
```
