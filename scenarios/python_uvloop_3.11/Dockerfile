ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_uvloop_3.11/main.py \
    ./scenarios/python_uvloop_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*


# Set the working directory to the location of the program
WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="5"
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false
ENV DD_TRACE_DEBUG=true
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"
ENV EXECUTION_TIME_SEC="5"

# Run the program when the container starts
CMD python main.py

