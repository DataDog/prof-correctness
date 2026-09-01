ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

# ddtrace stays the S3 pin from the base image. Do not pip-install
# requirements.txt (it includes ddtrace). uvloop has no cp315 wheel;
# compile the sdist.
RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential \
    && pip install uvloop \
    && apt-get purge -y build-essential \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

COPY ./scenarios/python_uvloop/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=5
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false

CMD python main.py
