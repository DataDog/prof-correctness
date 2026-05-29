ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_gil_contention_3.11/main.py \
    ./scenarios/python_gil_contention_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="30"

ENV DD_PROFILING_ENABLED="true"

CMD python main.py
