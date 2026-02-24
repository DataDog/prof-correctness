ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /app
COPY ./scenarios/python_cpu/ /usr/src/app

RUN chmod 644 /usr/src/app/main.py
# Install any needed packages specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC 30
# Run your python script when the container launches
ENV DD_TRACE_DEBUG=false
ENV DD_PROFILING_MEMORY_ENABLED=false

CMD ddtrace-run python main.py
