# full_host_otlp_cpu scenario
#
# Full-host CPU profiling with the standalone Datadog full-host profiler,
# exported over OTLP and captured to .otlp files (read natively by the
# analyzer). CPU counterpart to the pprof `full_host` scenario.
#
# The profiler + the generic otlp_dump capture sidecar come from the
# prof-fullhost-otlp base image (base_images/Dockerfile.fullhost-otlp); this
# scenario only adds the workload, config and start script.
ARG BASE_IMAGE="prof-fullhost-otlp"

# ----- build the CPU workload -----
FROM ubuntu:22.04 AS workload-build
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
ADD ./scenarios/full_host_otlp_cpu/workload.c .
# Binary name "cpu_workload" is what the assertion matches (frames are mapping
# basenames without symbol upload).
RUN gcc -O2 -fno-omit-frame-pointer -o cpu_workload workload.c

# ----- final image: base (profiler + otlp_dump) + our workload -----
FROM ${BASE_IMAGE} AS final
USER root
RUN mkdir -p /app/data

COPY --from=workload-build /src/cpu_workload /app/cpu_workload
ADD ./scenarios/full_host_otlp_cpu/host-profiler-config.yaml /app/host-profiler-config.yaml
ADD ./scenarios/full_host_otlp_cpu/start.sh /app/start.sh
RUN chmod 755 /app/start.sh

ENV EXECUTION_TIME_SEC="20"

# Override the base image's default host-profiler entrypoint with our
# orchestration. Requires privileges (provided by the harness for full_host*).
ENTRYPOINT []
CMD ["/app/start.sh"]
