FROM python:3.11

# When set, pre-install ddtrace from this URL (pip --find-links compatible).
# Used by upstream dd-trace-py CI to test with just-built wheels.
ARG DDTRACE_FIND_LINKS=""
RUN if [ -n "$DDTRACE_FIND_LINKS" ]; then \
      pip install --no-cache-dir --find-links "$DDTRACE_FIND_LINKS" --pre ddtrace; \
    fi

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
