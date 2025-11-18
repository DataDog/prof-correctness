ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

RUN pip download ddtrace --no-deps --no-index --find-links https://dd-trace-py-builds.s3.amazonaws.com/82657378/index.html
RUN pip install ddtrace-*.whl

# Copy the Python target into the container
COPY ./scenarios/python_asyncio_3.11/main.py /app/
RUN chmod 644 /app/*

ENV EXECUTION_TIME_SEC="2"

# Set the working directory to the location of the program
WORKDIR /app

# Run the program when the container starts
CMD python main.py
