import os
import time

from ddtrace.profiling import Profiler


class Target:
    def __init__(self) -> None:
        self.memory: list[bytearray] = []

    def run(self, n: int) -> None:
        end_time = time.monotonic() + n
        while time.monotonic() < end_time:
            self.allocate_memory(1024)
            self.allocate_memory_2(1024)

    def allocate_memory(self, size: int) -> None:
        self.memory.append(bytearray(size))

    def allocate_memory_2(self, size: int) -> None:
        self.memory.append(bytearray(3 * size))


if __name__ == "__main__":
    # Simple application that creates two threads with different durations:
    # - MainThread runs target() for 2 seconds
    # - Worker Thread-1 runs target() for 1 second
    # The profiler should capture both threads with their respective durations.
    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    Target().run(EXECUTION_TIME_SEC)

    prof.stop()
