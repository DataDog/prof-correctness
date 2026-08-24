from gevent import monkey

monkey.patch_all()

import os  # noqa: E402
import time  # noqa: E402
from threading import Thread  # noqa: E402


def target(n: float) -> None:
    end_time = time.monotonic() + n
    count = 0
    while time.monotonic() < end_time:
        count += 1
        if count % 1000 == 0:
            time.sleep(0.01)


def main() -> None:
    execution_time_sec = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    threads: list[Thread] = [Thread(target=target, args=(execution_time_sec / 2,)) for _ in range(10)]
    for thread in threads:
        thread.start()

    target(float(execution_time_sec))

    for thread in threads:
        thread.join()


if __name__ == "__main__":
    main()
