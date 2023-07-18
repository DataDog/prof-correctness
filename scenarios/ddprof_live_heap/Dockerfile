FROM ubuntu:20.04 AS base

RUN apt-get update \
    && apt-get install -y \
    xz-utils \
    gcc \
    g++ \
    curl \
    jq \
    wget \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app
WORKDIR /app

ADD ./scenarios/ddprof_live_heap/main.cc .
RUN g++ ./main.cc -o test

FROM base AS final
# Install native profiling
ADD ./profilers/ddprof/install_profiler.sh .
RUN ./install_profiler.sh /usr/local/bin

ENV EXECUTION_TIME="12"
# Default is that test data is dropped in the data folder
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_NATIVE_LOG_MODE="/app/data/ddprof_log"
# If upload period is > EXECUTION_TIME, it can cause issues when shutting down containers 
ENV DD_PROFILING_UPLOAD_PERIOD="10"
# ddprof -l debug -e "sALLOC period=1 mode=l" -u 30 ./test
# Force deterministic by sampling all allocations
CMD ddprof -l notice -e "sALLOC period=1 mode=l" /app/test
