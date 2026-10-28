# CPU stack gate (3.15 candidate)

Pair: `python_cpu_3.14` / `python_cpu_3.15`. Same workload and expectations as 3.14.

**Wheel-only** — requires `DDTRACE_INSTALL_URL`; excluded from `main` CI.

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh"
TEST_SCENARIOS='python_cpu_3\.15' go test -v -run TestScenarios
```
