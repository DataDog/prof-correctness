ARG BASE_IMAGE="prof-ddprof-ub-20.04"
FROM $BASE_IMAGE

COPY ./scenarios/ddprof_allocations/ ./
# Build the app using CMake
RUN cmake -S . -B build
RUN cmake --build build -- -j $(nproc)

CMD ["/app/run_ddprof.sh", "-l", "notice", "/app/build/simple_malloc" ]
