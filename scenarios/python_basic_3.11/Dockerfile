ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_basic_3.11/main.py \
    ./scenarios/python_basic_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*


# Set the working directory to the location of the program
WORKDIR /app

RUN pip download ddtrace --no-deps --no-index --find-links "https://dd-trace-py-builds.s3.amazonaws.com/5612418466de9ca4c9b4511b3e223934d231ecc0/index.html"
RUN pip install ddtrace-*.whl

ENV EXECUTION_TIME_SEC="2"
ENV DD_TRACE_DEBUG=true

# Run the program when the container starts
CMD python main.py
# CMD ddprof -l notice --preset cpu_live_heap python main.py
