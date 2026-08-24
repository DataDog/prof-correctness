# Shared exceptions gate scenario image. BASE_IMAGE is resolved from the scenario
# folder suffix (_3.14 / _3.15) when not set inline; see docker_helpers_test.go.
ARG BASE_IMAGE
FROM $BASE_IMAGE

COPY ./scenarios/python_gate/requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt

COPY ./scenarios/python_exceptions/main.py /app/main.py
WORKDIR /app

ENV DD_PROFILING_EXCEPTION_ENABLED=true

CMD python main.py
