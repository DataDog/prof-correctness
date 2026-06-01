ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_idle_baseline/main.py \
    ./scenarios/python_idle_baseline/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="10"

ENV DD_PROFILING_ENABLED="true"

ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED="true"

# 1% of the app's own CPU time.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_TARGET_OVERHEAD="1"

# Slowest allowed sampling interval (100ms). Without the baseline below, an idle
# app would back off to this floor and barely sample.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US="100000"

# Short stable-CPU window so the idle period dominates p_stable quickly (the
# startup CPU ages out within ~1s). With p_stable ~0 the only thing keeping the
# sampler alive is the baseline floor below.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_P_STABLE_WINDOW_S="1"
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_P_STABLE_PERCENTILE="95"

# The baseline floor under test: a minimum CPU budget that keeps the sampler
# running even when the app is idle.
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_BASELINE="5"

CMD python main.py
