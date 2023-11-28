ARG BASE_IMAGE="prof-ddprof-ub-20.04"
FROM $BASE_IMAGE

ADD ./scenarios/ddprof_live_heap/main.cc .
RUN g++ ./main.cc -o test

# Force deterministic by sampling all allocations
CMD /app/run_ddprof.sh -l notice -e "sALLOC period=1 mode=l" /app/test
