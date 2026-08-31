ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_gate/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_deep_stack/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=10
ENV DD_PROFILING_ENABLED=true

CMD python main.py
