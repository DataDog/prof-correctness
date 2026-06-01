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
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    end = time() + execution_time

    barrier = threading.Barrier(NUM_WORKERS)
    threads: list[threading.Thread] = [
        threading.Thread(target=worker, args=(barrier, end), name=f"worker-{i}") for i in range(NUM_WORKERS)
    ]

    for t in threads:
        t.start()

    for t in threads:
        t.join()
