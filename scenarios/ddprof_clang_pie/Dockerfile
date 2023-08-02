FROM ubuntu:20.04 AS base
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

ADD ./scenarios/ddprof_clang_pie/main.c .
ADD ./scenarios/ddprof_clang_pie/CMakeLists.txt .

# Build the app using CMake
RUN cmake -S . -B build -DCMAKE_LINKER=lld-${CLANG_VERSION}
RUN cmake --build build -- -j $(nproc)
# Checking the elf layout
RUN readelf -lW build/clang_pie

FROM base AS final
# Install native profiling
ADD ./profilers/ddprof/install_profiler.sh .
RUN ./install_profiler.sh /usr/local/bin

ENV EXECUTION_TIME_SEC="12"
# Default is that test data is dropped in the data folder
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_NATIVE_LOG_MODE="/app/data/ddprof_log"
# If upload period is > EXECUTION_TIME_SEC, it can cause issues when shutting down containers 
ENV DD_PROFILING_UPLOAD_PERIOD="10"
CMD ["ddprof", "-l", "notice", "/app/build/clang_pie" ]
