ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_gate/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_native_cpu/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=25
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0

CMD python main.py
