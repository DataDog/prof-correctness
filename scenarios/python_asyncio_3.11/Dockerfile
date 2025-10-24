ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

RUN pip install ddtrace

# Copy the Python target into the container
COPY ./scenarios/python_asyncio_3.11/main.py /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

# Run the program when the container starts
CMD python main.py
