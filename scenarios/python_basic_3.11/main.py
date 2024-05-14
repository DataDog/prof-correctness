import os
from threading import Thread
from time import sleep
from ddtrace.profiling import Profiler


def target(n):
    sleep(n)

if __name__ == "__main__":
    # Simple application that runs for two seconds. Each of the spawned threads
    # including the ones from the profiler will be attributed one second of
    # wall time for each second that the application runs. Except for the thread
    # that explicitly sleeps for only a second.
    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    t = Thread(target=target, args=(EXECUTION_TIME_SEC / 2,))
    t.start()

    target(EXECUTION_TIME_SEC)

    t.join()
