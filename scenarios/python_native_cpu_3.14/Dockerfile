ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_native_cpu_3.14/main.py \
    ./scenarios/python_native_cpu_3.14/requirements.txt \
    /app/

WORKDIR /app
RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC=10
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0

CMD python main.py
