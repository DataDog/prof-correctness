ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_thread_churn/main.py \
    ./scenarios/python_thread_churn/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="10"

ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED="true"
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US="10000"

ENV DD_PROFILING_ENABLED="true"

CMD python main.py
