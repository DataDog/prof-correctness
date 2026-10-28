import os
from time import time


def empty_method() -> None:
    pass


def slow_method() -> None:
    execution_time_sec = float(os.getenv("EXECUTION_TIME_SEC", "10"))
    end = time() + execution_time_sec
    while time() < end:
        x = "h" + "e" + "l" + "l" + "o" + ","
        x += "w" + "o" + "r" + "l" + "d"
        empty_method()


if __name__ == "__main__":
    slow_method()
