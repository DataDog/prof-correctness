FROM node:18-bullseye

# Give everyone write permissions in /app because profile is written here
RUN mkdir /app && chmod a+rwX /app
WORKDIR /app

COPY ./scenarios/node_heap/* ./
RUN chmod 755 /app/*
RUN npm install

ENV EXECUTION_TIME_SEC="10"
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_EXPORTERS=file
ENV DD_PROFILING_ENABLED=1
ENV DD_TRACING_ENABLED=0
ENV DD_REMOTE_CONFIGURATION_ENABLED=0
ENV DD_PROFILING_PROFILERS=space
ENV DD_TRACE_DEBUG=1
CMD node --expose-gc -r dd-trace/init main.js
