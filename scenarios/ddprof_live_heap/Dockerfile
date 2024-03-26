ARG BASE_IMAGE="prof-ddprof-ub-20.04"
FROM $BASE_IMAGE

ADD ./scenarios/ddprof_live_heap/main.cc .
RUN g++ ./main.cc -o test

CMD /app/run_ddprof.sh -l notice --preset cpu_live_heap /app/test
