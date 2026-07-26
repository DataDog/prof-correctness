ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_gevent_3.14/main.py \
    ./scenarios/python_gevent_3.14/requirements.txt \
    /app/

WORKDIR /app
RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC=2
ENV DD_PROFILING_ENABLED=true

CMD ddtrace-run python main.py
