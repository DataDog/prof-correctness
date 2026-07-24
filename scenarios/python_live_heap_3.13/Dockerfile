ARG BASE_IMAGE="prof-python-3.13"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_live_heap_3.13/main.py \
    ./scenarios/python_live_heap_3.13/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install --no-cache-dir -r requirements.txt

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

# Run the program when the container starts
CMD python main.py
