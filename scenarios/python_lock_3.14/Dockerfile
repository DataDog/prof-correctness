ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_lock_3.14/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_lock_3.14/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=10
ENV DD_PROFILING_LOCK_ENABLED=true

CMD python main.py
