# CPU stack gate (3.14 baseline)

Pair: `python_cpu_3.14` / `python_cpu_3.15`. Asserts `cpu-time` on CPU-bound loops with 2:1 share (`a` / `b`) and `thread name` labels. Memory profiling off.

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| cpu-time | `main;.*b` | 66 | 5 |
| cpu-time | `main;.*a` | 33 | 5 |

`scale_by_duration`: true. Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_cpu_3\.14' go test -v -run TestScenarios
```
