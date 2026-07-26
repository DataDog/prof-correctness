ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

COPY ./scenarios/python_uvloop_3.15/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=5
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false

CMD python main.py
