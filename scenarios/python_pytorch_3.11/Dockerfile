ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

WORKDIR /usr/src/app

COPY ./scenarios/python_pytorch_3.11/ /usr/src/app

RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC 30
ENV DD_TRACE_DEBUG=false
ENV DD_PROFILING_MEMORY_ENABLED=false

CMD ddtrace-run python main.py
