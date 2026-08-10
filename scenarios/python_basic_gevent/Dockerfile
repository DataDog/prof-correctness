ARG BASE_IMAGE="prof-python-3.12"
FROM $BASE_IMAGE

ENV EXECUTION_TIME_SEC="2"

# Copy the Python target into the container
COPY ./scenarios/python_basic_gevent/main.py \
    ./scenarios/python_basic_gevent/requirements.txt \
    /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN pip install -r requirements.txt

# Run the program when the container starts
CMD ddtrace-run python main.py
