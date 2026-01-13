# Use an official Python runtime as a parent image
FROM python:3.11

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /app
COPY ./scenarios/python_cpu/ /usr/src/app

RUN chmod 644 /usr/src/app/main.py
# Install any needed packages specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC 30
# Run your python script when the container launches
ENV DD_PROFILING_ENABLED true
ENV DD_TRACE_ENABLED false
ENV DD_TRACE_DEBUG false
ENV DD_PROFILING_MEMORY_ENABLED=false
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"

CMD ddtrace-run python main.py
