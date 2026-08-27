from __future__ import annotations

import contextlib
import os
import time

from ddtrace.profiling import Profiler


def raise_value_error() -> None:
    msg: str = "prof-correctness exception sample"
    raise ValueError(msg)


def handle_value_error() -> None:
    with contextlib.suppress(ValueError):
        raise_value_error()


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()
    execution_time_raw: str = os.environ.get("EXECUTION_TIME_SEC", "3")
    end: float = time.monotonic() + float(execution_time_raw)
    while time.monotonic() < end:
        handle_value_error()
    prof.stop()


if __name__ == "__main__":
    main()
