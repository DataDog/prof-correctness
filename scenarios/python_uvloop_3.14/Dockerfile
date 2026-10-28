ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_uvloop/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_uvloop/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=5
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false

CMD python main.py
