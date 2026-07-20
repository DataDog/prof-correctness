ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

RUN pip install ddtrace

COPY ./scenarios/python_async_gen_3.14/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=8
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true

CMD ddtrace-run python main.py
