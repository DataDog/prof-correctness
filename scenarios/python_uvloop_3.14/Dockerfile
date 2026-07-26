ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_uvloop_3.14/main.py \
    ./scenarios/python_uvloop_3.14/requirements.txt \
    /app/

WORKDIR /app
RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC=5
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false

CMD python main.py
