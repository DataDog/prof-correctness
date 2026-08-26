import contextlib
import os
import time

from ddtrace.profiling import Profiler


def raise_value_error() -> None:
    msg = "prof-correctness exception sample"
    raise ValueError(msg)


def handle_value_error() -> None:
    with contextlib.suppress(ValueError):
        raise_value_error()


def main() -> None:
    prof = Profiler()
    prof.start()
    end = time.monotonic() + float(os.environ.get("EXECUTION_TIME_SEC", "3"))
    while time.monotonic() < end:
        handle_value_error()
    prof.stop()


if __name__ == "__main__":
    main()
