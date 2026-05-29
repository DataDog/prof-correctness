import os
import threading
from time import sleep, time


def busy_loop(end: float) -> None:
    x = 0
    while time() < end:
        for i in range(10000):
            x += i


def idle_wait(end: float) -> None:
    while time() < end:
        sleep(0.05)


if __name__ == "__main__":
    from ddtrace.profiling import Profiler

    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    end = time() + execution_time

    busy = threading.Thread(target=busy_loop, args=(end,), name="busy")
    idle = threading.Thread(target=idle_wait, args=(end,), name="idle")
    busy.start()
    idle.start()
    busy.join()
    idle.join()
