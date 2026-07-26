ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

# ddtrace is pre-installed in the base image when DDTRACE_INSTALL_URL is set
# (dd-trace-py downstream CI). Do not pip-install from PyPI here: the persistent
# live-heap profile is recent and may not be in a published release yet, so this
# scenario runs against the just-built wheel until it ships.
COPY ./scenarios/python_live_heap_3.14/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC="15"
# Enable the live-heap profiler and make the snapshot the deliverable.
ENV DD_PROFILING_MEMORY_ENABLED=true
ENV DD_PROFILING_HEAP_ENABLED=true
# Sample interval == object size -> ~one sample per object -> stable
# per-stack proportions for both heap-space and heap-live-samples.
ENV DD_PROFILING_HEAP_SAMPLE_SIZE=16384
# Export several snapshots during the run so the persistent live set is
# exercised across upload intervals (not just at shutdown).
ENV DD_PROFILING_UPLOAD_INTERVAL=5
# Isolate the heap profiler so other collectors don't add noise.
ENV DD_PROFILING_STACK_ENABLED=false
ENV DD_PROFILING_LOCK_ENABLED=false

CMD python main.py
