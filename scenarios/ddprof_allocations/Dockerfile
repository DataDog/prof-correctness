FROM ubuntu:20.04 AS base
ENV GCC_VERSION=10
ENV CC=gcc-${GCC_VERSION}
ENV CXX=g++-${GCC_VERSION}


RUN apt-get update && apt-get install -y \
    g++-${GCC_VERSION} \
    gcc-${GCC_VERSION} \
    cmake \
    curl \
    xz-utils \
    jq \
    wget \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app
WORKDIR /app

COPY ./scenarios/ddprof_allocations/ ./
# Build the app using CMake
RUN cmake -S . -B build
RUN cmake --build build -- -j $(nproc)

# Install native profiling
ARG CACHE_DATE=2023-03-01_09:58:27
COPY ./binaries/ /app/binaries/
ADD ./profilers/ddprof/install_profiler.sh .
RUN ./install_profiler.sh /usr/local/bin

ENV EXECUTION_TIME="11"
# Default is that test data is dropped in the data folder
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_UPLOAD_PERIOD="10"
# One allocation of size 1000 every millisecond 
CMD ["ddprof", "-l", "notice", "/app/build/simple_malloc" ]
