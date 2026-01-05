from gevent import monkey

monkey.patch_all()

import os  # noqa: E402
import time  # noqa: E402
from threading import Thread  # noqa: E402


def target(n: int) -> None:
    # Do actual work instead of just sleeping so profiler can capture it
    end_time = time.monotonic() + n
    count = 0
    while time.monotonic() < end_time:
        count += 1
        if count % 1000 == 0:
            # Yield to gevent
            time.sleep(0.01)


if __name__ == "__main__":
    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    threads = [Thread(target=target, args=(EXECUTION_TIME_SEC / 2,)) for _ in range(10)]
    for thread in threads:
        thread.start()

    target(EXECUTION_TIME_SEC)

    for thread in threads:
        thread.join()
