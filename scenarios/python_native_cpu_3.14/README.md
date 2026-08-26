# Native CPU gate (3.14 baseline)

Pair: `python_native_cpu_3.14` / `python_native_cpu_3.15`. Asserts **cpu-time** and **wall-time** attribute CPU to Python frames calling C-extension stdlib functions (hashlib, zlib, math, re, binascii). Workload: `scenarios/python_native_cpu/main.py`. Adaptive sampling disabled.

| Profile type | Stack | Expected | Margin |
|--------------|-------|----------|--------|
| cpu-time | `hash_work` | 5000000000 | 5 |
| cpu-time | `compress_work` | 5000000000 | 5 |
| cpu-time | `factorial_work` | 5000000000 | 5 |
| cpu-time | `regex_work` | 5000000000 | 5 |
| cpu-time | `crc_work` | 5000000000 | 5 |
| wall-time | `hash_work` | 5000000000 | 5 |
| wall-time | `compress_work` | 5000000000 | 5 |
| wall-time | `factorial_work` | 5000000000 | 5 |
| wall-time | `regex_work` | 5000000000 | 5 |
| wall-time | `crc_work` | 5000000000 | 5 |

**Source:** five equal slices of `EXECUTION_TIME_SEC/5` = 5e9 ns. Matching sum 25e9 ns per profile type. `factorial_work` uses `math.factorial(10_000)` so one unit is small vs the time window. Runs on `main` CI (PyPI ddtrace).

```sh
TEST_SCENARIOS='python_native_cpu_3\.14' go test -v -run TestScenarios
```
