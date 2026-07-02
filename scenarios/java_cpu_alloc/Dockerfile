# Multi-stage scenario for the Datadog java-profiler.
#
# Stage 1 builds jfr2pprof (btraceio/jafar, PROF-15286) from source, pinned to a
# known-good commit, since it is not yet published as a Maven Central artifact
# or a GitHub release asset.
# Stage 2 downloads the java-profiler native agent from a GitHub release,
# compiles the workload, runs it under the agent to record a .jfr, then
# converts that recording to pprof via jfr2pprof + datadog.yaml (PROF-15289).
FROM eclipse-temurin:25-jdk AS jfr2pprof-build

RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

# jfr2pprof's `parser` module dependency declares a Java 8 toolchain and
# jfr2pprof itself declares Java 21; jafar's settings.gradle doesn't wire up
# the foojay-resolver plugin, so Gradle can't auto-provision either -
# install both explicitly and point Gradle at them below.
RUN ARCH="$(dpkg --print-architecture)" \
    && case "$ARCH" in \
         amd64) JDK_ARCH=x64 ;; \
         arm64) JDK_ARCH=aarch64 ;; \
         *) echo "unsupported arch: $ARCH" >&2; exit 1 ;; \
       esac \
    && mkdir -p /opt/jdk8 /opt/jdk21 \
    && curl -fsSL "https://api.adoptium.net/v3/binary/latest/8/ga/linux/${JDK_ARCH}/jdk/hotspot/normal/eclipse" \
       | tar xz -C /opt/jdk8 --strip-components=1 \
    && curl -fsSL "https://api.adoptium.net/v3/binary/latest/21/ga/linux/${JDK_ARCH}/jdk/hotspot/normal/eclipse" \
       | tar xz -C /opt/jdk21 --strip-components=1

WORKDIR /build
# Pinned to the jafar release that fixes https://github.com/btraceio/jafar/issues/105
# (frame.format was parsed but never applied to pprof Function.Name).
ARG JAFAR_COMMIT=8dde4267210952b2e3347c72efd4cd54833e903d
RUN git init . \
    && git remote add origin https://github.com/btraceio/jafar.git \
    && git fetch --depth 1 origin ${JAFAR_COMMIT} \
    && git checkout FETCH_HEAD

RUN ./gradlew --no-daemon \
      -Porg.gradle.java.installations.paths=/opt/jdk8,/opt/jdk21 \
      -Porg.gradle.java.installations.auto-detect=false \
      -Porg.gradle.java.installations.auto-download=false \
      :jfr2pprof:shadowJar \
    && cp jfr2pprof/build/libs/jfr2pprof-*-all.jar /build/jfr2pprof.jar

FROM eclipse-temurin:25-jdk AS workload-build

COPY ./scenarios/java_cpu_alloc/Workload.java /build/Workload.java
RUN mkdir -p /build/classes && javac -d /build/classes /build/Workload.java

FROM eclipse-temurin:25-jre AS runtime

WORKDIR /app

# Datadog java-profiler native agent (must match a released libjavaProfiler.so).
ARG JAVA_PROFILER_VERSION=1.45.0
RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/* \
    && ARCH="$(dpkg --print-architecture)" \
    && case "$ARCH" in \
         amd64) SO_ARCH=x64 ;; \
         arm64) SO_ARCH=arm64 ;; \
         *) echo "unsupported arch: $ARCH" >&2; exit 1 ;; \
       esac \
    && curl -fsSL -o /app/libjavaProfiler.so \
       "https://github.com/DataDog/java-profiler/releases/download/v_${JAVA_PROFILER_VERSION}/libjavaProfiler_linux-${SO_ARCH}.so"

COPY --from=jfr2pprof-build /build/jfr2pprof.jar /app/jfr2pprof.jar
COPY --from=workload-build /build/classes /app
COPY ./scenarios/java_cpu_alloc/datadog.yaml /app/datadog.yaml
COPY ./scenarios/java_cpu_alloc/run.sh /app/run.sh

RUN chmod 755 /app/run.sh

ENV EXECUTION_TIME_SEC="10"

CMD ["/app/run.sh"]
