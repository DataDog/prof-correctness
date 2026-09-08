# Multi-stage scenario for the Datadog java-profiler.
#
# Stage 1 compiles the workload.
# Stage 2 downloads the java-profiler native agent from a GitHub release,
# runs the workload under the agent to record a .jfr, then converts that
# recording to pprof via jfr2pprof (btraceio/jafar, launched through jbang's
# jafar-tools-latest@btraceio alias) + datadog.yaml.
FROM eclipse-temurin:25-jdk AS workload-build

COPY ./scenarios/java_cpu_alloc/CpuAllocWorkload.java /build/CpuAllocWorkload.java
RUN mkdir -p /build/classes && javac -d /build/classes /build/CpuAllocWorkload.java

FROM eclipse-temurin:25-jdk AS runtime

WORKDIR /app

# Datadog java-profiler native agent. Defaults to the latest GitHub release;
# pass --build-arg JAVA_PROFILER_TAG=v_1.45.0 to pin a specific candidate
# (e.g. to bisect a regression), matching how the other prof-correctness
# profilers (ddprof's latest-rc tag, dd-trace-py/-rb/-js's unpinned/master
# dependencies) default to latest with an explicit pin as the escape hatch.
ARG JAVA_PROFILER_TAG=""
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && ARCH="$(dpkg --print-architecture)" \
    && case "$ARCH" in \
         amd64) SO_ARCH=x64 ;; \
         arm64) SO_ARCH=arm64 ;; \
         *) echo "unsupported arch: $ARCH" >&2; exit 1 ;; \
       esac \
    && TAG="${JAVA_PROFILER_TAG:-$(curl -fsSL https://api.github.com/repos/DataDog/java-profiler/releases/latest | grep -m1 '"tag_name"' | cut -d'"' -f4)}" \
    && echo "Using java-profiler release ${TAG}" \
    && curl -fsSL -o /app/libjavaProfiler.so \
       "https://github.com/DataDog/java-profiler/releases/download/${TAG}/libjavaProfiler_linux-${SO_ARCH}.so"

# jfr2pprof, via jbang's jafar-tools-latest@btraceio alias (btraceio/jbang-catalog).
# Resolved at build time so run.sh doesn't need network access; jbang re-checks
# for a newer snapshot on every `jbang app install`/first run of a given
# coordinate but is otherwise offline once cached under JBANG_DIR.
# JBANG_DIR is pinned to a fixed, world-writable path (rather than the
# default ~/.jbang) because the container is run as an arbitrary non-root
# uid:gid (see prof-correctness's -u flag), which has no /etc/passwd entry
# and thus no writable $HOME. HOME is likewise repointed here since jbang's
# Maven resolver falls back to $HOME/.m2 regardless of JBANG_DIR.
ENV JBANG_DIR="/opt/jbang"
ENV HOME="${JBANG_DIR}"
RUN apt-get update && apt-get install -y --no-install-recommends unzip \
    && rm -rf /var/lib/apt/lists/* \
    && curl -Ls https://sh.jbang.dev | bash -s - app setup \
    && chmod -R a+rwX "${JBANG_DIR}"
ENV PATH="${JBANG_DIR}/bin:${PATH}"
# jfr2pprof has no --help; running it bare still resolves/caches the jbang
# dependency, it just exits non-zero after printing its usage string.
RUN jbang trust add https://github.com/btraceio/ \
    && (jbang jafar-tools-latest@btraceio jfr2pprof || true) \
    && chmod -R a+rwX "${JBANG_DIR}"

COPY --from=workload-build /build/classes /app
COPY ./scenarios/java_cpu_alloc/datadog.yaml /app/datadog.yaml
COPY ./scenarios/java_cpu_alloc/run.sh /app/run.sh

RUN chmod 755 /app/run.sh

ENV EXECUTION_TIME_SEC="10"

CMD ["/app/run.sh"]
