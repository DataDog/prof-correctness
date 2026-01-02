ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_greenlet_multithreaded_3.11/requirements.txt /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install --no-cache-dir -r requirements.txt

COPY ./scenarios/python_greenlet_multithreaded_3.11/main.py /app/

ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false
ENV DD_TRACE_DEBUG=false
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"
ENV EXECUTION_TIME_SEC="10"

CMD ddtrace-run python main.py
