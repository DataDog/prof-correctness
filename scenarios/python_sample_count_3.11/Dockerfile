ARG BASE_IMAGE="prof-python-3.11"
FROM $BASE_IMAGE

COPY ./scenarios/python_sample_count_3.11/main.py \
    ./scenarios/python_sample_count_3.11/requirements.txt \
    /app/
RUN chmod 644 /app/*

WORKDIR /app

RUN pip install -r requirements.txt

ENV EXECUTION_TIME_SEC="30"

ENV DD_PROFILING_ENABLED="true"

ENV _DD_PROFILING_STACK_ADAPTIVE_SAMPLING_ENABLED="false"

# Enable fast memory copying
ENV _DD_PROFILING_STACK_FAST_COPY="true"

CMD python main.py
