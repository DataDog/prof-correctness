FROM php:8.2.5-cli

RUN mkdir /app
WORKDIR /app

RUN apt-get update && apt-get install -y \
    jq \
    && rm -rf /var/lib/apt/lists/*

COPY ./profilers/php/install_profiler.sh .
RUN ./install_profiler.sh 

COPY ./scenarios/php_time/main.php .
RUN chmod 644 ./main.php

ENV EXECUTION_TIME_SEC="10"
ENV DD_PROFILING_OUTPUT_PPROF="/app/data/php.pprof"
ENV DD_PROFILING_EXPERIMENTAL_CPU_TIME_ENABLED=true
ENV DD_PROFILING_EXPERIMENTAL_ALLOCATION_ENABLED=false
ENV DD_PROFILING_ENABLED=true
CMD [ "php", "/app/main.php" ]
