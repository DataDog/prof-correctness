ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_many_threads_3.11/main.py \
    ./scenarios/python_many_threads_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="30"

ENV DD_PROFILING_ENABLED="true"

# Force reservoir sampling: there are 20 worker threads but at most 5 are
# sampled per cycle, so the inverse-probability wall-time weighting must
# reconstruct the correct totals.
# We could go down to 1 or 2 but it makes the test very flaky. Even with
# 5/20 sampling and 15% margin, the point is sufficiently proven.
ENV _DD_PROFILING_STACK_MAX_THREADS="5"

# Reservoir sampling with two threads results in very noisy wall time for
# many threads. By reducing the interval, we get more samples and less noise.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=1
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US="10000"

CMD python main.py
