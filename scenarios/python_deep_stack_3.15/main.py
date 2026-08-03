import os
import sys
from time import time

from ddtrace.profiling import Profiler

DEPTH = 400


def burn(end: float) -> None:
    x = 0
    while time() < end:
        for i in range(10000):
            x += i


def recurse(depth: int, end: float) -> None:
    if depth <= 0:
        burn(end)
        return
    recurse(depth - 1, end)


if __name__ == "__main__":
    sys.setrecursionlimit(10000)

    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    end = time() + execution_time
    recurse(DEPTH, end)
