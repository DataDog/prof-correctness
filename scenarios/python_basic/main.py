import os
from threading import Thread
from time import sleep


def target(n):
    sleep(n)


if __name__ == "__main__":
    execution_time = int(os.environ.get("EXECUTION_TIME", "2"))

    t = Thread(target=target, args=(execution_time / 2,))
    t.start()

    target(execution_time)

    t.join()
