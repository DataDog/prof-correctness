# Python downstream gate (dd-trace-py)

Paired **3.14 (baseline)** and **3.15 (candidate)** prof-correctness scenarios
exercise the Python profiling stack for the 3.14 → 3.15 migration. They are the
intended default set when dd-trace-py triggers downstream CI on profiling changes.

**Scenarios land in follow-up PRs** (core families first, then feature-specific
pairs). This directory is an index only — not a runnable scenario.

## Scenarios

| Family | 3.14 (baseline) | 3.15 (candidate) | PR |
|--------|-----------------|---------------------|-----|
| _(pending)_ | — | — | Core scenarios in follow-up PRs |

## Default downstream regexp

Once scenarios are added, dd-trace-py should pass an explicit regexp (not the
downstream workflow default of `python.*`). The regexp grows as families merge;
see each PR for the current value.

Override via `workflow_dispatch` → `test_scenarios`, or when triggering
[`downstream-python.yml`](../../.github/workflows/downstream-python.yml) manually.

## Base images

Gate scenarios use [`base_images/Dockerfile.python-wheel`](../../base_images/Dockerfile.python-wheel)
with a `PYTHON_IMAGE` build arg (`python:3.14` / `python:3.15.0b1`) and optional
`DDTRACE_INSTALL_URL` for wheel pre-install.

## Wheel install

Every scenario builds against a **dd-trace-py wheel** via `DDTRACE_INSTALL_URL`
(as `downstream-python.yml` does:
`https://dd-trace-py-builds.s3.amazonaws.com/<sha>/install.sh`), pre-installed in
the base image.

- **All `*_3.15` folders** — PyPI wheels may not be published for 3.15 yet;
  excluded from prof-correctness `main` CI today (see `test_scenarios_exclude` in
  [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)).
- **Wheel-only 3.14 folders** — as they land, add them to `test_scenarios_exclude`
  in `ci.yml` until the feature ships on PyPI (not excluded in #165 — no such
  scenarios yet).

## Local run

```sh
export DDTRACE_INSTALL_URL="https://dd-trace-py-builds.s3.amazonaws.com/<commit-sha>/install.sh"
TEST_SCENARIOS='<gate-regexp>' go test -v -run TestScenarios
```

## Gate lifecycle

This gate tests the **migration delta** (3.14 → 3.15). It is time-boxed: retire
the paired 14v15 framing at 3.15 GA and fold workloads into steady-state
prof-correctness on {oldest, newest} supported Python versions.

## Further reading

- prof-correctness downstream wiring: [README](../../README.md#downstream-from-dd-trace-py)
