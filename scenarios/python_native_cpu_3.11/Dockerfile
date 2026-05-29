ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_native_cpu_3.11/main.py \
    ./scenarios/python_native_cpu_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="25"
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0

CMD python main.py
