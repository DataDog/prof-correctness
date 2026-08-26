# Exceptions gate (3.14 baseline)

Pair: `python_exceptions_3.14` / `python_exceptions_3.15`. Workload: `scenarios/python_exceptions/main.py` (ValueError loop for `EXECUTION_TIME_SEC`).

| Profile type | Stack | Expected % | Margin | Labels |
|--------------|-------|------------|--------|--------|
| exception-samples | `.*raise_value_error.*` | 100 | 20 | exception type: `builtins.ValueError` |

**Source:** only workload exception; target 100% of exception-samples, margin is non-workload frames.

```sh
TEST_SCENARIOS='python_exceptions_3\.14' go test -v -run TestScenarios
```
