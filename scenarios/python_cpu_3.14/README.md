# CPU stack gate (3.14 baseline)

Pair: `python_cpu_3.14` / `python_cpu_3.15`. Asserts `cpu-time` on CPU-bound loops with 2:1 share (`a` / `b`) and `thread name` labels. Memory profiling off.

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| cpu-time | `<module>;.*CPUBurner\.main;CPUBurner\.b` | 66 | 5 |
| cpu-time | `<module>;.*CPUBurner\.main;CPUBurner\.a` | 33 | 5 |

**Source:** CI burn-in on PyPI ddtrace ([#169](https://github.com/DataDog/prof-correctness/pull/169)). Two `CPUBurner` loops run with 2:1 iteration ratio (`b` twice as long as `a`); percents are empirical wall/cpu samples, not exact math from loop counts alone.

`scale_by_duration`: true. Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_cpu_3\.14' go test -v -run TestScenarios
```
