ARG BASE_IMAGE="prof-python-3.12"
FROM $BASE_IMAGE

WORKDIR /usr/src/app

COPY ./scenarios/python_cpu_sleep_sync_3.12/ /usr/src/app

RUN chmod 644 /usr/src/app/main.py
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC=30
ENV DD_PROFILING_MEMORY_ENABLED=false
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0

CMD ddtrace-run python main.py
