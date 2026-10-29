ARG BASE_IMAGE="prof-python-3.12"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_basic_memory_3.12/requirements.txt /app/requirements.txt
RUN chmod 644 /app/*

RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_basic_memory_3.12/main.py /app/main.py

# Set the working directory to the location of the program
WORKDIR /app

ENV DD_TRACE_DEBUG=false
# Allocation counts are driven by samples on small objects. At the default 1 MiB interval,
# this run gets about 130 of them and count shares vary by 5+ points. 64 KiB gives 16x more.
ENV DD_PROFILING_HEAP_SAMPLE_SIZE=65536

# Run the program when the container starts
CMD python main.py
# CMD ddprof -l notice --preset cpu_live_heap python main.py
