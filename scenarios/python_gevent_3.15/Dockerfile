ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

# ddtrace stays the S3 pin from the base image. Do not pip-install
# requirements.txt (it includes ddtrace). gevent ships cp315 wheels.
RUN pip install gevent

COPY ./scenarios/python_gevent/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=2
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true

CMD ddtrace-run python main.py
