from ddtrace.profiling import Profiler


def raise_value_error() -> None:
    raise ValueError("prof-correctness exception sample")


def handle_value_error() -> None:
    try:
        raise_value_error()
    except ValueError:
        pass


if __name__ == "__main__":
    prof = Profiler()
    prof.start()
    for _ in range(500):
        handle_value_error()
    prof.stop()
