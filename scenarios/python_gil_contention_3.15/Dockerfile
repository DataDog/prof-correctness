ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

COPY ./scenarios/python_gil_contention_3.15/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=15
ENV DD_PROFILING_ENABLED=true

CMD python main.py
