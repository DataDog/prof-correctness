ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_mem_domain_3.14/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_mem_domain_3.14/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=15
ENV DD_PROFILING_HEAP_ENABLED=true
ENV DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true
ENV DD_PROFILING_HEAP_SAMPLE_SIZE=524288
ENV DD_PROFILING_UPLOAD_INTERVAL=5
ENV DD_PROFILING_STACK_ENABLED=false
ENV DD_PROFILING_LOCK_ENABLED=false

CMD python main.py
