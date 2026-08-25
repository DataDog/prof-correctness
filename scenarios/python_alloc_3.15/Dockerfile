ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

# ddtrace is pre-installed in the base image when DDTRACE_INSTALL_URL is set
# (dd-trace-py downstream CI). Do not pip-install here — PyPI wheels may not
# exist for 3.15 yet.

COPY ./scenarios/python_alloc/main.py /app/main.py
WORKDIR /app

ENV DD_PROFILING_STACK_ENABLED=false
ENV DD_PROFILING_LOCK_ENABLED=false

CMD python main.py
