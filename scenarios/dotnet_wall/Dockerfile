# Use the official .NET SDK image to build and run the application
FROM mcr.microsoft.com/dotnet/sdk:7.0 AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the project files into the container
COPY ./scenarios/dotnet_wall .

# Build the application
RUN dotnet publish -c Release -o out

# Use the official .NET runtime image to run the application
FROM mcr.microsoft.com/dotnet/runtime:7.0

# Set the working directory inside the container
WORKDIR /app

# Copy the build output into the runtime image
COPY --from=build /app/out .

# Copy the queries.txt file into the runtime image
COPY --from=build /app/queries.txt .

# this forces docker to not use the cache for the next commands
ARG CACHE_DATE=2023-03-01_09:58:27
#
# Download and install Datadog .NET APM
RUN apt-get update && apt-get install -y curl jq
RUN URL_PACKAGE=$(curl -s https://api.github.com/repos/DataDog/dd-trace-dotnet/releases/latest | jq -r '.assets[] | select(.name | startswith("datadog-dotnet-apm")).browser_download_url' | grep tar.gz | grep -v musl | grep -v arm) \
    && curl -L -o datadog-dotnet-apm.tar.gz $URL_PACKAGE
RUN mkdir -p /opt/datadog
RUN tar -C /opt/datadog -xzf datadog-dotnet-apm.tar.gz
RUN /opt/datadog/createLogPath.sh

# Set environment variables for Datadog
ENV CORECLR_ENABLE_PROFILING=1 \
    CORECLR_PROFILER="{846F5F1C-F9AE-4B07-969E-05C26BC060D8}" \
    CORECLR_PROFILER_PATH="/opt/datadog/Datadog.Trace.ClrProfiler.Native.so" \
    DD_DOTNET_TRACER_HOME="/opt/datadog" \
    DD_INTEGRATIONS="/opt/datadog/integrations.json" \
    DD_SERVICE="prof-correctness-dotnet-workload" \
    DD_ENV="production" \
    DD_VERSION="1.0"
ENV LD_PRELOAD=/opt/datadog/continuousprofiler/Datadog.Linux.ApiWrapper.x64.so

ENV DD_INTERNAL_PROFILING_OUTPUT_DIR=/app/data
ENV DD_PROFILING_ENABLED=1
ENV DD_TRACE_DEBUG=TRUE

# Set the entry point for the container
# ENTRYPOINT ["dotnet", "WorkloadScheduler.dll"]
# ENTRYPOINT [ "bash" ]
CMD dotnet WorkloadScheduler.dll
