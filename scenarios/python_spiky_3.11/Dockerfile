ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_spiky_3.11/main.py \
    ./scenarios/python_spiky_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="30"

ENV DD_PROFILING_ENABLED="true"

ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED="true"
 
# 1% of the app's own CPU time
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_TARGET_OVERHEAD="1"

# Maximum interval between stack samples in microseconds (max allowed is 100ms)
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_MAX_INTERVAL_US="100000"

# 1% CPU regardless of what the app is using
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_BASELINE="1"

# Lookback window of 60 seconds for CPU usage
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_P_STABLE_WINDOW_S="60"

# Use the 95th percentile of the CPU usage over the window to determine CPU usage
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_P_STABLE_PERCENTILE="95"

# Enable fast memory copying
ENV _DD_PROFILING_STACK_FAST_COPY="true"

CMD python main.py
