ARG BASE_IMAGE="prof-python-3.12"
FROM $BASE_IMAGE

COPY ./scenarios/python_mem_domain_3.12/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_mem_domain_3.12/main.py /app/main.py
WORKDIR /app

ENV DD_PROFILING_HEAP_ENABLED=true
ENV DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED=true

CMD python main.py
