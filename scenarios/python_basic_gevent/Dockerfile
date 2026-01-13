ARG BASE_IMAGE="prof-python-3.10"
FROM $BASE_IMAGE

ENV EXECUTION_TIME_SEC="2"

# Copy the Python target into the container
COPY ./scenarios/python_basic_gevent/main.py \
    ./scenarios/python_basic_gevent/requirements.txt \
    /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN pip download ddtrace --no-deps --no-index --find-links https://dd-trace-py-builds.s3.amazonaws.com/90138225/index.html
RUN if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
RUN pip uninstall ddtrace || echo "ddtrace not installed"
RUN pip install ddtrace-*.whl


# Run the program when the container starts
CMD ddtrace-run python main.py
