ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

WORKDIR /usr/src/app

COPY ./scenarios/python_safe_point_bias_3.11/ /usr/src/app

RUN chmod 644 /usr/src/app/main.py
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC=20
ENV DD_TRACE_DEBUG=false
ENV DD_PROFILING_MEMORY_ENABLED=false

CMD ddtrace-run python main.py
