# Shared mem-domain gate scenario image. BASE_IMAGE is resolved from the scenario
# folder suffix (_3.14 / _3.15) when not set inline; see docker_helpers_test.go.
ARG BASE_IMAGE
FROM $BASE_IMAGE

# ddtrace is pre-installed in the base image when DDTRACE_INSTALL_URL is set
# (dd-trace-py downstream CI). Do not pip-install from PyPI here: MEM-domain
# default-on is not in stable 4.14.1, so this scenario runs against the
# just-built / pinned S3 wheel.
COPY ./scenarios/python_mem_domain/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC="15"
ENV DD_PROFILING_MEMORY_ENABLED=true
ENV DD_PROFILING_HEAP_ENABLED=true
# Pin sample interval so the Poisson window does not follow host RAM
# (max(ceil(total_mem/65536), 1 MiB)). 512 KiB matches the dd-trace-py
# MEM-domain unit test (16 MiB alloc → sampled with weight ≈ BUFFER_SIZE).
ENV DD_PROFILING_HEAP_SAMPLE_SIZE=524288
# Isolate the heap profiler so other collectors don't add noise.
ENV DD_PROFILING_STACK_ENABLED=false
ENV DD_PROFILING_LOCK_ENABLED=false
# Do not pin DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED: this gate exercises
# the default-on path (True on dd-trace-py main).

CMD python main.py
