import os
import threading
from time import time

from ddtrace.profiling import Profiler

NUM_WORKERS = 20


def worker(barrier: threading.Barrier, end: float) -> None:
    # Make sure no thread starts before all threads are ready
    barrier.wait()

    x = 0
    while time() < end:
        for i in range(10000):
            x += i


if __name__ == "__main__":
    prof = Profiler()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    end = time() + execution_time

    barrier = threading.Barrier(NUM_WORKERS + 1)
    threads: list[threading.Thread] = [
        threading.Thread(target=worker, args=(barrier, end), name=f"worker-{i}") for i in range(NUM_WORKERS)
    ]

    for t in threads:
        t.start()

    prof.start()

    # Every worker is now alive and parked on the barrier, so the profiler has
    # registered all 20 threads before any work begins. Release them together
    # so no thread misses early sampling cycles.
    barrier.wait()

    for t in threads:
        t.join()
