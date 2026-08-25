# Asyncio task labels gate (3.15 candidate)

Pair: `python_asyncio_3.14` / `python_asyncio_3.15`. Same workload and expectations as 3.14 (`scenarios/python_asyncio/main.py`, `python_asyncio/profile.json`).

**Wheel-only** — requires `DDTRACE_INSTALL_URL`; excluded from `main` CI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_asyncio_3\.15' go test -v -run TestScenarios
```
