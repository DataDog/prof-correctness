ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

COPY ./scenarios/python_native_cpu_3.15/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=10
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0

CMD python main.py
