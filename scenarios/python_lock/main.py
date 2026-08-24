import os
import threading
import time

from ddtrace.profiling import Profiler


def lock_churn(lock: threading.Lock, end: float) -> None:
    while time.time() < end:
        with lock:
            pass


def main() -> None:
    prof = Profiler()
    prof.start()

    lock = threading.Lock()
    execution_time_sec = float(os.getenv("EXECUTION_TIME_SEC", "10"))
    end = time.time() + execution_time_sec

    workers: list[threading.Thread] = [threading.Thread(target=lock_churn, args=(lock, end)) for _ in range(2)]
    for worker in workers:
        worker.start()
    for worker in workers:
        worker.join()

    prof.stop()


if __name__ == "__main__":
    main()
