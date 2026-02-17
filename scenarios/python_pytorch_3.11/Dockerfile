FROM python:3.11

WORKDIR /usr/src/app

COPY ./scenarios/python_pytorch_3.11/ /usr/src/app

RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC 30
ENV DD_PROFILING_ENABLED true
ENV DD_TRACE_ENABLED false
ENV DD_TRACE_DEBUG false
ENV DD_PROFILING_MEMORY_ENABLED=false
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"

CMD ddtrace-run python main.py
