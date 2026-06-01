import os
import threading
from time import time

from ddtrace.profiling import Profiler


def churn_worker() -> None:
    x = 0
    for _ in range(50):
        for i in range(5_000):
            x += i


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "15"))
    end = time() + execution_time

    while time() < end:
        start = time()
        t = threading.Thread(target=churn_worker, name="churn")
        t.start()
        t.join()
