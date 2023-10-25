ARG BASE_IMAGE="prof-python-3.10"
FROM $BASE_IMAGE

# Copy the Python target into the container
COPY ./scenarios/python_basic_3.10/main.py \
    ./scenarios/python_basic_3.10/requirements.txt \
    /app/
RUN chmod 644 /app/*


# Set the working directory to the location of the program
WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="2"

# Run the program when the container starts
CMD python main.py
# CMD ddprof -l notice --preset cpu_live_heap python main.py
