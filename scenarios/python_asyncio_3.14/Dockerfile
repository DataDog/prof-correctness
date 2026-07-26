ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_asyncio_3.14/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_asyncio_3.14/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=5
ENV ECHION_USE_FAST_COPY_MEMORY=1
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true

CMD ddtrace-run python main.py
