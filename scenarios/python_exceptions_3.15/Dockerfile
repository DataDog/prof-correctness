ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

COPY ./scenarios/python_exceptions_3.15/main.py /app/main.py
WORKDIR /app

ENV DD_PROFILING_EXCEPTION_ENABLED=true

CMD python main.py
