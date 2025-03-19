# Use Rocky Linux 9.3 as base image
FROM rockylinux:9.3

# Install Temurin JDK 11.0.20.1+1
RUN dnf install -y \
    glibc-langpack-en \
    && dnf clean all \
    && curl -LO https://github.com/adoptium/temurin11-binaries/releases/download/jdk-11.0.20.1%2B1/OpenJDK11U-jdk_x64_linux_hotspot_11.0.20.1_1.tar.gz \
    && tar -xzf OpenJDK11U-jdk_x64_linux_hotspot_11.0.20.1_1.tar.gz -C /opt/ \
    && rm OpenJDK11U-jdk_x64_linux_hotspot_11.0.20.1_1.tar.gz

# Set JAVA_HOME environment variable
ENV JAVA_HOME=/opt/jdk-11.0.20.1+1
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


ENV DD_PROFILING_DDPROF_ENABLED=true
# Run the application with Datadog profiler
CMD ["java", "-javaagent:/dd-java-agent.jar", "CpuBurner"]
