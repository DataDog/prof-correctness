import os
from time import sleep, time

from ddtrace.profiling import Profiler


def idle_wait(end: float) -> None:
    while time() < end:
        sleep(0.05)


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "10"))
    idle_wait(time() + execution_time)
