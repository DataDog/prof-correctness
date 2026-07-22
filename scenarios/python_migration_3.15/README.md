# Python 3.15 migration scenarios

These four scenarios mirror the `*_3.14` migration gate set but require a
**dd-trace-py wheel** baked into the base image at build time:

- Set `DDTRACE_INSTALL_URL` when running `go test -run TestScenarios` (as
  `downstream-python.yml` does for dd-trace-py PRs).
- Do **not** `pip install ddtrace` in the scenario Dockerfile — PyPI may not
  publish 3.15 wheels yet.

They are intentionally excluded from prof-correctness `main` CI until 3.15
wheels are generally available; they run in the dd-trace-py downstream gate.
