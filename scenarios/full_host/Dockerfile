# todo use this : docker pull ghcr.io/datadog/dd-otel-host-profiler:v0.4.0
# add an env var to the service being run, check that we have the split by service.

FROM ubuntu:20.04 AS build
ENV CLANG_VERSION=11
ENV CC=clang-${CLANG_VERSION}
ENV CXX=clang++-${CLANG_VERSION}

RUN apt-get update && apt-get install -y \
    clang-${CLANG_VERSION} \
    lld-${CLANG_VERSION} \
    cmake \
    curl \
    jq \
    wget \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app
WORKDIR /app

ADD ./scenarios/full_host/main.c .
ADD ./scenarios/full_host/CMakeLists.txt .

# Build the app using CMake
RUN cmake -S . -B build -DCMAKE_LINKER=lld-${CLANG_VERSION}
RUN cmake --build build -- -j $(nproc)
# Checking the elf layout
RUN readelf -lW build/clang_pie

FROM ubuntu:20.04 AS final

# Install required dependencies for the full host profiler
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Download and install dd-otel-host-profiler from main
# https://github.com/DataDog/dd-otel-host-profiler/releases/download/main/dd-otel-host-profiler-main-x86_64.tar.gz
RUN curl -L https://github.com/DataDog/dd-otel-host-profiler/releases/download/main/dd-otel-host-profiler-main-x86_64.tar.gz -o dd-otel-host-profiler.tar.gz && \
    tar -xzf dd-otel-host-profiler.tar.gz && \
    mv dd-otel-host-profiler /usr/local/bin/ && \
    chmod +x /usr/local/bin/dd-otel-host-profiler && \
    rm dd-otel-host-profiler.tar.gz

# Copy the built application
COPY --from=build /app/build/clang_pie /app/clang_pie

# Create data directory for pprof output
RUN mkdir -p /app/data

# Copy and set up the startup script
ADD ./scenarios/full_host/start.sh /app/start.sh
RUN chmod 755 /app/start.sh

ENV EXECUTION_TIME_SEC="12"

# The profiler needs higher privileges to access system-level profiling
# This will be handled by docker run --privileged or --cap-add SYS_ADMIN

CMD ["/app/start.sh"]
