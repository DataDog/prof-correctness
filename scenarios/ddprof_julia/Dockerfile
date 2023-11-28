FROM julia:latest

RUN mkdir /app
RUN mkdir /app/binaries

COPY ./binaries/ /app/binaries/
COPY ./scenarios/ddprof_julia/*.jl /app/
RUN chmod 644 /app/*.jl

WORKDIR /app

RUN apt-get update && apt-get install -y \
    xz-utils \
    jq \
    wget \
    bash \
    procps \
    && rm -rf /var/lib/apt/lists/*


# Install native profiling
ARG CACHE_DATE=2023-03-01_09:58:27
ADD ./profilers/ddprof/install_profiler.sh .
RUN ./install_profiler.sh /usr/local/bin
ADD ./profilers/ddprof/run_ddprof.sh .

ENV EXECUTION_TIME_SEC="11"
# Default is that test data is dropped in the data folder
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
# Julia uses LLVM symbol maps (in .debug folder)
ENV ENABLE_JITPROFILING=1
# Why we have the LD_LIBRARY_PATH updated
# Issue with how we intercept dlopen
# The rpath is defined by the caller of dlopen. Intercepting the call changes
# who is calling dlopen. This is not easy to fix, so I'm adding the path to Julia libs
# Debug command:
# ENABLE_JITPROFILING=1 LD_LIBRARY_PATH=~/dd/scripts/julia-1.8.5/lib/julia/ EXECUTION_TIME_SEC=30 ddprof -l notice 
# --show_sample julia scenarios/ddprof_julia/main.jl
ENV LD_LIBRARY_PATH="/usr/local/julia/lib:${LD_LIBRARY_PATH}"
RUN mkdir /.debug && mkdir /.debug/jit && chmod 777 /.debug/jit
CMD ["/app/run_ddprof.sh", "-l", "notice", "--preset", "cpu_only", "julia", "main.jl" ]
