ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

WORKDIR /usr/src/app

COPY ./scenarios/python_pytorch_profiler_3.11/ /usr/src/app

RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC=30
ENV DD_TRACE_DEBUG=false
ENV DD_PROFILING_MEMORY_ENABLED=false
# Exercise the torch profiler integration (torch.profiler.profile is monkey-patched).
ENV DD_PROFILING_PYTORCH_ENABLED=true
# Disable the Python stack sampler so the cpu-time profile contains only the
# torch-operator samples emitted by the integration. This makes the per-operator
# percentages stable and lets us assert on the reconstructed operator nesting.
ENV DD_PROFILING_STACK_ENABLED=false

ENV USER=app
CMD ddtrace-run python main.py
