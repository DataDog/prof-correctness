FROM eclipse-temurin:21-jdk

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install the Datadog Java tracer, which carries the Datadog native Java
# profiler (ddprof).  The profiler writes Datadog-specific JFR events such as
# datadog.ExecutionSample; this scenario validates that prof-correctness can
# consume those recordings directly.
RUN curl -fsSL -o /app/dd-java-agent.jar https://dtdg.co/latest-java-tracer

COPY ./scenarios/java_cpu_jfr/DummyApp.java .
RUN javac DummyApp.java

# EXECUTION_TIME_SEC is injected by the test harness; default to 60 s.
ENV EXECUTION_TIME_SEC=60

# Run the app under the Datadog Java profiler and use the tracer's built-in
# debug dump path to persist the JFR snapshot locally. The prof-correctness
# harness finds the timestamped dd-profiler-debug-*.jfr file by regex.
CMD sh -eu -c '\
    mkdir -p /app/data/dumps /app/data/tmp; \
    java \
      -javaagent:/app/dd-java-agent.jar \
      -Ddd.service=prof-correctness-java-cpu-jfr \
      -Ddd.env=local \
      -Ddd.trace.enabled=false \
      -Ddd.profiling.enabled=true \
      -Ddd.profiling.ddprof.enabled=true \
      -Ddd.profiling.ddprof.cpu.enabled=true \
      -Ddd.profiling.ddprof.wall.enabled=false \
      -Ddd.profiling.upload.period=30 \
      -Ddd.profiling.upload.timeout=1 \
      -Ddd.profiling.start-force-first=true \
      -Ddd.profiling.tempdir=/app/data/tmp \
      -Ddd.profiling.debug.dump_path=/app/data/dumps \
      -Ddd.profiling.url=http://127.0.0.1:8126/ \
      -Ddd.telemetry.enabled=false \
      -Ddd.jmxfetch.enabled=false \
      DummyApp \
'
