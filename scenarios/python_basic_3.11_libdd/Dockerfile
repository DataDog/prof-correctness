ARG BASE_IMAGE="prof-python-3.11-libdd"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_basic_3.11_libdd/main.py \
    ./scenarios/python_basic_3.11_libdd/requirements.txt \
    /app/
RUN chmod 644 /app/*


# Set the working directory to the location of the program
WORKDIR /app

RUN pip install --no-cache-dir -r requirements.txt

ENV EXECUTION_TIME_SEC="2"

# Run the program when the container starts
CMD python main.py
# CMD ddprof -l notice --preset cpu_live_heap python main.py

