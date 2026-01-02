ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_basic_memory_3.11/requirements.txt /app/requirements.txt
RUN chmod 644 /app/*

RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_basic_memory_3.11/main.py /app/main.py

# Set the working directory to the location of the program
WORKDIR /app

ENV DD_TRACE_DEBUG=false

# Run the program when the container starts
CMD python main.py
# CMD ddprof -l notice --preset cpu_live_heap python main.py
