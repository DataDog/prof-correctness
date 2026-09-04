ARG BASE_IMAGE="prof-python-3.15"
FROM $BASE_IMAGE

# Parked: uvloop has no cp315 wheel. The sdist died SIGSEGV (pip exit -11)
# getting requirements on 3.15. Do not pip-install requirements.txt
# (it includes ddtrace and would overwrite the S3 pin).
COPY ./scenarios/python_uvloop/main.py /app/main.py
WORKDIR /app

ENV EXECUTION_TIME_SEC=5
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false

CMD python main.py
