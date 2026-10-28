ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_gate/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_alloc/main.py /app/main.py
WORKDIR /app

ENV DD_PROFILING_STACK_ENABLED=false
ENV DD_PROFILING_LOCK_ENABLED=false
# Pin sample interval so the rate does not follow host RAM
# (max(ceil(total_mem/65536), 1 MiB)). 65536 is above ~31 KB so
# TRACEBACK_ARRAY_MAX_COUNT (65535) is not the limiter even if a
# future change reintroduces retention of the 1M objects.
ENV DD_PROFILING_HEAP_SAMPLE_SIZE=65536
# bytearray buffers are PYMEM_DOMAIN_MEM on 3.13+; without this hook
# the gate only sees ~56-byte OBJ headers (~80 MB total).
ENV DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true

CMD python main.py
