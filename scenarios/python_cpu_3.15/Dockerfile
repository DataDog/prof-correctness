# Shared CPU gate scenario image (3.15 candidate). BASE_IMAGE is resolved from the
# scenario folder suffix (_3.14 / _3.15) when not set inline; see docker_helpers_test.go.
ARG BASE_IMAGE
FROM $BASE_IMAGE

# ddtrace is pre-installed in the base image when DDTRACE_INSTALL_URL is set
# (dd-trace-py downstream CI). Do not pip-install here — PyPI wheels may not
# exist for 3.15 yet.

COPY ./scenarios/python_cpu_gate/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=10
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true
ENV DD_PROFILING_MEMORY_ENABLED=false

CMD ddtrace-run python main.py
