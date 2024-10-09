ARG BASE_IMAGE="prof-python-3.11-libdd"
FROM $BASE_IMAGE

# Install Rust
# RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
# ENV PATH="/root/.cargo/bin:${PATH}"

# Copy the Python target into the container
COPY ./scenarios/python_basic_3.11_v2/main.py \
    ./scenarios/python_basic_3.11_v2/requirements.txt \
    /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

# ENV CMAKE_BUILD_PARALLEL_LEVEL=16
RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC="2"
ENV DD_PROFILING_STACK_V2_ENABLED="1"

# Run the program when the container starts
CMD python main.py
# CMD ddprof -l notice --preset cpu_live_heap python main.py

