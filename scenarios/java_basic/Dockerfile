# Use OpenJDK 18 as base image
FROM openjdk:18-jdk

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

# Copy Java application to the container
WORKDIR /app
COPY ./scenarios/java_basic/CpuBurner.java .

# Compile Java application
RUN javac CpuBurner.java

# Run the application with Datadog profiler
CMD ["java", "-javaagent:/dd-java-agent.jar", "CpuBurner"]
