import os
import threading
import time

from ddtrace.profiling import Profiler

NUM_THREADS = 8
NO_YIELD_TIME_MS = 100


def spin(end: float) -> None:
    x = 0
    next_yield_time = time.time() + (NO_YIELD_TIME_MS / 1000)
    while time.time() < end:
        for i in range(10_000):
            x += i

        # Help the scheduler yield to other threads
        if time.time() > next_yield_time:
            time.sleep(0.0001)
            next_yield_time = time.time() + (NO_YIELD_TIME_MS / 1000)


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    end = time.time() + execution_time

    threads: list[threading.Thread] = [
        threading.Thread(target=spin, args=(end,), name=f"spin-{i}") for i in range(NUM_THREADS)
    ]
    for t in threads:
        t.start()
    for t in threads:
        t.join()
