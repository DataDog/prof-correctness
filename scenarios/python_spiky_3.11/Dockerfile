ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_spiky_3.11/main.py \
    ./scenarios/python_spiky_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="20"

# Leave empty to use ddtrace defaults.
# _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_TARGET_OVERHEAD controls the fraction of
# CPU (0-100) the profiler is allowed to consume; lower means sparser sampling.
# _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US is the hard upper bound
# on the interval between stack samples in microseconds; higher means sparser sampling.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_TARGET_OVERHEAD="1"
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US="10000"

CMD python main.py
