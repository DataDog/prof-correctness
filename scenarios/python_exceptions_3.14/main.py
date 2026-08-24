import contextlib

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
    for _ in range(500):
        handle_value_error()
    prof.stop()


if __name__ == "__main__":
    main()
