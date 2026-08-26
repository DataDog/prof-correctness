# uvloop gate (3.14 baseline)

Pair: `python_uvloop_3.14` / `python_uvloop_3.15`. Workload: `scenarios/python_uvloop/main.py` (uvloop + two named sequential tasks, no gather).

| Profile type | Stack | Expected | Margin | Labels |
|--------------|-------|----------|--------|--------|
| wall-time | `.*cpu_bound_work.*` | 2500000000 | 10 | task name: `cpu_task` |
| wall-time | `.*io_simulation.*` | 2500000000 | 10 | task name: `io_task` |

**Source:** each task runs `EXECUTION_TIME_SEC/2` = 2.5e9 ns. Matching sum 5e9 ns. Main is never a gather parent and never a leaf.

```sh
TEST_SCENARIOS='python_uvloop_3\.14' go test -v -run TestScenarios
```
