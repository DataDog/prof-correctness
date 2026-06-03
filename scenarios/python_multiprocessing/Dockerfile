ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_multiprocessing/main.py \
    ./scenarios/python_multiprocessing/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="10"

ENV DD_PROFILING_ENABLED="true"

# Denser sampling so the short run yields enough samples per process.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=1
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US="10000"

CMD python main.py
