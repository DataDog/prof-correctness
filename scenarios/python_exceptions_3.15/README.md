# Exceptions gate (3.15 candidate)

Pair: `python_exceptions_3.14` / `python_exceptions_3.15`. Same workload and expectations as 3.14 (`scenarios/python_exceptions/main.py`, `profile.json`).

Expectation tables: see `python_exceptions_3.14/README.md`. 3.15 not re-burned on main CI until manylinux cp315 wheel lands.

**Wheel-only** — requires `DDTRACE_INSTALL_URL`; excluded from `main` CI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_exceptions_3\.15' go test -v -run TestScenarios
```
