ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

# ddtrace stays the S3 pin from the base image. Do not pip-install
# requirements.txt (it includes ddtrace). No cp315 wheel; build from sdist.
RUN pip install --no-binary :all: uvloop

COPY ./scenarios/python_uvloop/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=5
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false

CMD python main.py
