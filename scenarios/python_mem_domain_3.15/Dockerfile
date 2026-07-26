ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

# ddtrace is pre-installed in the base image when DDTRACE_INSTALL_URL is set
# (dd-trace-py downstream CI). Do not pip-install here — PyPI wheels may not
# exist for 3.15 yet.

COPY ./scenarios/python_mem_domain_3.15/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=15
ENV DD_PROFILING_HEAP_ENABLED=true
ENV DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true
ENV DD_PROFILING_HEAP_SAMPLE_SIZE=524288
ENV DD_PROFILING_UPLOAD_INTERVAL=5
ENV DD_PROFILING_STACK_ENABLED=false
ENV DD_PROFILING_LOCK_ENABLED=false

CMD python main.py
