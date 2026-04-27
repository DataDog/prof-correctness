FROM node:18-bullseye

# Give everyone write permissions in /app because profile is written here
RUN mkdir /app && chmod a+rwX /app
WORKDIR /app

COPY ./scenarios/node_heap_oom/* ./
RUN chmod 755 /app/*
RUN npm install

ENV EXECUTION_TIME_SEC="10"
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_EXPORTERS=file
ENV DD_PROFILING_ENABLED=1
ENV DD_TRACING_ENABLED=0
# TEMPORARY: Since dd-trace-js PR #6982, DD_TRACING_ENABLED=0 also disables the
# entire tracer module (NoopProxy), which prevents the profiler from starting.
# DD_DYNAMIC_INSTRUMENTATION_ENABLED=true is on the allowlist (added in PR #7916)
# that keeps the real proxy alive while tracing stays off. Remove once dd-trace-js
# adds DD_PROFILING_ENABLED to that allowlist.
ENV DD_DYNAMIC_INSTRUMENTATION_ENABLED=true
ENV DD_REMOTE_CONFIGURATION_ENABLED=0
# Disable telemetry so its beforeExit HTTP request doesn't allocate thousands
# of objects that dominate the heap snapshot and skew percentage assertions.
ENV DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
ENV DD_PROFILING_PROFILERS=space
ENV DD_PROFILING_EXPERIMENTAL_OOM_MONITORING_ENABLED=1
ENV DD_TRACE_DEBUG=1
CMD node --expose-gc --max-old-space-size=50 -r dd-trace/init main.js || true
