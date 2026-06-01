import multiprocessing
import os
from time import time

from ddtrace.profiling import Profiler

NUM_WORKERS = 4

# Assigned in __main__ and inherited (via fork) by the child processes.
prof: Profiler


def work(end: float) -> None:
    x = 0
    while time() < end:
        for i in range(10000):
            x += i


def run_worker(end: float) -> None:
    work(end)
    # multiprocessing exits children with os._exit(), which skips atexit, so the
    # profiler inherited across the fork would never flush. Stop it explicitly so
    # the child exports its own pprof file.
    prof.stop()


if __name__ == "__main__":
    # Fork is the interesting case: the children inherit the already-running
    # native sampler and must restart it via the pthread_atfork handler.
    multiprocessing.set_start_method("fork")

    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "10"))
    end = time() + execution_time

    procs: list[multiprocessing.Process] = [
        multiprocessing.Process(target=run_worker, args=(end,), name=f"child-{i}") for i in range(NUM_WORKERS)
    ]
    for p in procs:
        p.start()

    # The parent runs the same work so every process' profile is comparable.
    work(end)

    for p in procs:
        p.join()

    prof.stop()
