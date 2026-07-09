FROM eclipse-temurin:21-jdk

WORKDIR /app

# Copy and compile the Java source
COPY ./scenarios/java_cpu_jfr/DummyApp.java .
RUN javac DummyApp.java

# Force re-run of the next layer when the cache date changes.
ARG CACHE_DATE=2024-01-01

# EXECUTION_TIME_SEC is injected by the test harness; default to 30 s.
ENV EXECUTION_TIME_SEC=30

# Run with built-in JFR.
#   -XX:+FlightRecorder          : enable JFR (no-op on JDK 11+ where it's always on)
#   -XX:StartFlightRecording=... : start a continuous recording; dump on JVM exit.
#   settings=profile             : enable CPU/alloc/lock profiling at low overhead.
CMD java \
    -XX:+FlightRecorder \
    "-XX:StartFlightRecording=filename=/app/data/profile.jfr,dumponexit=true,settings=profile,name=correctness" \
    DummyApp
