# Python downstream gate (dd-trace-py)

Two prof-correctness scenarios exercise **MEM domain** heap profiling on **3.14
(baseline)** and **3.15 (candidate)**. This is **feature coverage** (off by
default: `DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED`), not default-profiler gate.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) |
|--------|-------------------|------------------|
| mem-domain | `python_mem_domain_3.14` | `python_mem_domain_3.15` |

## Default downstream regexp

```
python_mem_domain_3\.(14|15)
```

## Wheel install

- **3.15** — wheel-only (`DDTRACE_INSTALL_URL`); excluded from `main` CI.
- **3.14** — runs on `main` CI against PyPI ddtrace.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_mem_domain_3\.(14|15)' go test -v -run TestScenarios
```
