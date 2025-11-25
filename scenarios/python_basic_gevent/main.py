from gevent import monkey

monkey.patch_all()

import os  # noqa: E402
from threading import Thread  # noqa: E402


def target(n: int) -> None:
    # Do actual work instead of just sleeping so profiler can capture it
    import time

    end_time = time.monotonic() + n
    count = 0
    while time.monotonic() < end_time:
        count += 1
        if count % 1000 == 0:
            # Yield to gevent
            time.sleep(0.01)


if __name__ == "__main__":
    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    t = Thread(target=target, args=(EXECUTION_TIME_SEC / 2,))
    t.start()

    target(EXECUTION_TIME_SEC)

    t.join()
