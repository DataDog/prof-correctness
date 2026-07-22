ARG BASE_IMAGE="prof-python-3.14"
FROM $BASE_IMAGE

COPY ./scenarios/python_exceptions_3.14/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_exceptions_3.14/main.py /app/main.py
WORKDIR /app

ENV DD_PROFILING_EXCEPTION_ENABLED=true

CMD python main.py
