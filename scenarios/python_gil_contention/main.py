import os
import sys
import threading
import time

from ddtrace.profiling import Profiler

NUM_THREADS: int = 8
NO_YIELD_TIME_MS: int = 100


def spin(end: float) -> None:
    x: int = 0
    next_yield_time: float = time.monotonic() + (NO_YIELD_TIME_MS / 1000)
    while time.monotonic() < end:
        for i in range(10_000):
            x += i

        # Help the scheduler yield to other threads
        if time.monotonic() > next_yield_time:
            time.sleep(0.0001)
            next_yield_time = time.monotonic() + (NO_YIELD_TIME_MS / 1000)


def main() -> None:
    # Shorter GIL slices → more handoffs over the fixed window, so per-thread
    # cpu-time converges closer to the 100/NUM_THREADS fair share.
    sys.setswitchinterval(0.001)

    prof: Profiler = Profiler()
    prof.start()

    execution_time: float = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    end: float = time.monotonic() + execution_time

    threads: list[threading.Thread] = [
        threading.Thread(target=spin, args=(end,), name=f"spin-{i}") for i in range(NUM_THREADS)
    ]
    for t in threads:
        t.start()
    for t in threads:
        t.join()


if __name__ == "__main__":
    main()
