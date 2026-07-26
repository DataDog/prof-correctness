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

## Base images

Gate scenarios use [`base_images/Dockerfile.python-wheel`](../../base_images/Dockerfile.python-wheel)
with a `PYTHON_IMAGE` build arg (`python:3.14` / `python:3.15.0b1`) and optional
`DDTRACE_INSTALL_URL` for wheel pre-install.

## Wheel install

Every scenario builds against a **dd-trace-py wheel** via `DDTRACE_INSTALL_URL`
when downstream CI runs (`https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh`).

- **3.15** — wheel-only; excluded from prof-correctness `main` CI (see
  `test_scenarios_exclude` in [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)).
- **3.14** — runs on `main` CI against PyPI ddtrace.

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='python_mem_domain_3\.(14|15)' go test -v -run TestScenarios
```
