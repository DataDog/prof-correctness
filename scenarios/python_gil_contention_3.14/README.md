# GIL contention gate (3.14 baseline)

Pair: `python_gil_contention_3.14` / `python_gil_contention_3.15`. Workload: `scenarios/python_gil_contention/main.py` (8 threads spinning with periodic yields).

Expectations in `scenarios/python_gil_contention/profile.json` (`scale_by_duration: true` — values are rates × profile duration).

| Profile type | Stack | Expected | Margin | Notes |
|--------------|-------|----------|--------|-------|
| cpu-time | `.*spin` (aggregate) | 1×10⁹ | 8 | one core |
| cpu-time | `.*spin` | 1.25×10⁸ | 32 | thread name: `spin-0`; fair share 1e9/NUM_THREADS; margin is GIL unfairness (old 12±4pp band) |
| wall-time | `.*spin` (aggregate) | 8×10⁹ | 8 | NUM_THREADS × 1e9 |

```sh
TEST_SCENARIOS='python_gil_contention_3\.14' go test -v -run TestScenarios
```
