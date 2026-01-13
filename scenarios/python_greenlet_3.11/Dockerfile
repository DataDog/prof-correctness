ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_greenlet_3.11/requirements.txt /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN pip download ddtrace --no-deps --no-index --find-links https://dd-trace-py-builds.s3.amazonaws.com/90138225/index.html
RUN if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
RUN pip uninstall ddtrace || echo "ddtrace not installed"
RUN pip install ddtrace-*.whl


COPY ./scenarios/python_greenlet_3.11/main.py /app/

ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_ENABLED=false
ENV DD_TRACE_DEBUG=false
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"
ENV EXECUTION_TIME_SEC="4"

# Run the program when the container starts
CMD ddtrace-run python main.py
