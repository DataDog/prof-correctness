ARG BASE_IMAGE="prof-python-3.13"
FROM $BASE_IMAGE

# asyncio
# OK:     4.0.0 / 4.0.1 / 4.0.2 / 4.0.3 / 4.1.0 / 4.1.1 / 4.1.2 / 
# Not OK: 4.1.3

# non asyncio
# OK:     4.1.1
# Not OK: 4.1.2

# NOT OK: 4.1.0
# OK: 4.0.3

RUN pip install ddtrace==4.1.0

# RUN pip download ddtrace --no-deps --no-index --find-links https://dd-trace-py-builds.s3.amazonaws.com/89868759/index.html
# RUN pip install ddtrace-*.whl

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
