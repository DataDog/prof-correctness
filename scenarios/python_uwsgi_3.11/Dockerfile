ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_uwsgi_3.11/requirements.txt /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN pip install -r requirements.txt

COPY ./scenarios/python_uwsgi_3.11/app.py /app/
COPY ./scenarios/python_uwsgi_3.11/uwsgi.ini /app/

ENV DD_TRACE_ENABLED=false
ENV DD_PROFILING_ENABLED=true
ENV DD_TRACE_DEBUG=false
ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED=0
ENV ECHION_USE_FAST_COPY_MEMORY=1
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/profiles"

# It's important that those are roughly the same, but execution time is slightly longer so
# that we're sure the Profile has been dumped.
ENV EXECUTION_TIME_SEC=11
ENV DD_PROFILING_UPLOAD_INTERVAL=10

# Run the program when the container starts
CMD ["uwsgi", "--ini", "uwsgi.ini"]
