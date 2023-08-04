import os
from threading import Thread
from time import sleep


def target(n):
    sleep(n)


if __name__ == "__main__":
    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    t = Thread(target=target, args=(EXECUTION_TIME_SEC / 2,))
    t.start()

    target(EXECUTION_TIME_SEC)

    t.join()
