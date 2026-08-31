import os
import sys
from time import time
from typing import Final

from ddtrace.profiling import Profiler

DEPTH: Final[int] = 400


def burn(end: float) -> None:
    x: int = 0
    while time() < end:
        for i in range(10000):
            x += i


def recurse(depth: int, end: float) -> None:
    if depth <= 0:
        burn(end)
        return
    recurse(depth - 1, end)


def main() -> None:
    sys.setrecursionlimit(10000)

    prof: Profiler = Profiler()
    prof.start()

    execution_time: float = float(os.environ.get("EXECUTION_TIME_SEC", "10"))
    end: float = time() + execution_time
    recurse(DEPTH, end)


if __name__ == "__main__":
    main()
