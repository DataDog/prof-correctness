FROM python:3.10

ENV EXECUTION_TIME="2"
ENV DD_PROFILING_ENABLED true
ENV DD_TRACE_ENABLED false
ENV DD_TRACE_DEBUG true
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"

# Copy the Python target into the container
COPY ./scenarios/python_basic/main.py \
    ./scenarios/python_basic/requirements.txt \
    /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN pip install -r requirements.txt

# Run the program when the container starts
CMD ddtrace-run python main.py
