ARG BASE_IMAGE="prof-ddprof-ub-20.04"
FROM $BASE_IMAGE

ADD ./scenarios/ddprof_inline/inline.cc .
ADD ./scenarios/ddprof_inline/build.sh .
RUN ./build.sh

CMD /app/run_ddprof.sh -l notice --inlined_functions true --show_config ./inline_test
