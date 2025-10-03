import os
from threading import Thread
from time import sleep
from ddtrace.profiling import Profiler


def target(n):
    sleep(n)

if __name__ == "__main__":
    # Simple application that creates two threads with different durations:
    # - MainThread runs target() for 2 seconds
    # - Worker Thread-1 runs target() for 1 second
    # The profiler should capture both threads with their respective durations.
    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    t = Thread(target=target, args=(EXECUTION_TIME_SEC / 2,))
    t.start()

    target(EXECUTION_TIME_SEC)

    t.join()
