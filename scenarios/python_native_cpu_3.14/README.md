# Native CPU gate (3.14 baseline)

Pair: `python_native_cpu_3.14` / `python_native_cpu_3.15`. Asserts **cpu-time** and **wall-time** attribute CPU to Python frames calling C-extension stdlib functions (hashlib, zlib, math, re, binascii). Workload: `scenarios/python_native_cpu/main.py`. Adaptive sampling disabled.

| Profile type | Stack | Expected % | Margin |
|--------------|-------|------------|--------|
| cpu-time | `hash_work` | 20 | 5 |
| cpu-time | `compress_work` | 20 | 5 |
| cpu-time | `factorial_work` | 20 | 5 |
| cpu-time | `regex_work` | 20 | 5 |
| cpu-time | `crc_work` | 20 | 5 |
| wall-time | `hash_work` | 20 | 5 |
| wall-time | `compress_work` | 20 | 5 |
| wall-time | `factorial_work` | 20 | 5 |
| wall-time | `regex_work` | 20 | 5 |
| wall-time | `crc_work` | 20 | 5 |

Runs on `main` CI (PyPI ddtrace).

**Anomaly:** five equal-budget functions sum to **96.5%** (range 95–98) in both cpu-time and wall-time across 70 CI runs. `hash_work`/`compress_work`/`crc_work` are exactly 19 in 70/70; `factorial_work` mean is 20.47. Hypothesis: `math.factorial(100_000)` overshoots its ~0.67s budget because the loop checks the clock then runs one whole unit. Investigate before tightening margins below ±5.

```sh
TEST_SCENARIOS='python_native_cpu_3\.14' go test -v -run TestScenarios
```
