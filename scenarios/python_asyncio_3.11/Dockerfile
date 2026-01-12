ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

RUN pip install ddtrace

# Copy the Python target into the container
COPY ./scenarios/python_asyncio_3.11/main.py /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

ENV DD_PROFILING_ENABLED=true
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV EXECUTION_TIME_SEC="3"
ENV DD_PROFILING_MEMORY_ENABLED=false
ENV DD_REMOTE_CONFIGURATION_ENABLED=false

# Run the program when the container starts
CMD ddtrace-run python main.py
