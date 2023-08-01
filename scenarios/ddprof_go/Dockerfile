# Dockerfile

FROM golang:1.20.5

WORKDIR /app

COPY ./scenarios/ddprof_go/ .

RUN go build main.go

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
CMD ["ddprof", "-l", "notice", "/app/main" ]
