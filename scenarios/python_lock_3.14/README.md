# Lock gate (3.14 baseline)

Pair: `python_lock_3.14` / `python_lock_3.15`. Workload: `scenarios/python_lock/main.py` (two threads contending on one `threading.Lock` via `lock_churn`).

| Profile type | Stack | Expected % | Margin | Labels |
|--------------|-------|------------|--------|--------|
| lock-acquire | `.*lock_churn.*` | 100 | 5 | lock name: `main.py:18:lock` (regex) |
| lock-release | `.*lock_churn.*` | 100 | 5 | lock name: `main.py:18:lock` (regex) |

**Source:** CI burn-in on PyPI ddtrace ([#166](https://github.com/DataDog/prof-correctness/pull/166)). 100% reflects all lock events attributed to the single shared lock; acquire and release profile types checked separately.

```sh
TEST_SCENARIOS='python_lock_3\.14' go test -v -run TestScenarios
```
