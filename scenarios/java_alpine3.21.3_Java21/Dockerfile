# Use Alpine latest as base image
FROM alpine:3.21.3


# Install dependencies and Temurin JDK 21.0.6+7
RUN apk add --no-cache \
    bash \
    curl \
    openjdk21 \
    && ln -sf /usr/lib/jvm/java-21-openjdk /opt/jdk-21.0.6+7

# Set JAVA_HOME environment variable
ENV JAVA_HOME=/opt/jdk-21.0.6+7
ENV PATH="${JAVA_HOME}/bin:${PATH}"

# Set environment variables for Datadog
ENV DD_SERVICE=cpu-burner-app
ENV DD_ENV=production
ENV DD_PROFILING_ENABLED=true
ENV DD_LOGS_INJECTION=true
ENV DD_TRACE_DEBUG=true
ENV DD_PROFILING_LOG_LEVEL=debug
ENV KEEP_JFRS=true

# Install Datadog Java agent
RUN curl -LO https://dtdg.co/latest-java-tracer && mv latest-java-tracer dd-java-agent.jar

# Set working directory
WORKDIR /app

# Copy Java application to the container
COPY ./scenarios/java_basic/CpuBurner.java .

# Compile Java application
RUN javac CpuBurner.java

# Run the application with Datadog profiler
CMD ["java", "-javaagent:/dd-java-agent.jar", "CpuBurner"]
