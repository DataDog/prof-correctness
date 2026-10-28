from ddtrace.profiling import Profiler


class Target:
    def __init__(self) -> None:
        self.memory: list[bytearray | None] = []
        self.index = 0
        self.grow_list(target=int(1e6))

    def run(self) -> None:
        while self.memory[-1] is None:
            self.allocate_memory_1(1024)
            self.allocate_memory_2(1024)

    def grow_list(self, target: int) -> None:
        self.memory = [None for _ in range(target)]

    def allocate_memory_1(self, size: int) -> None:
        self.memory[self.index] = bytearray(size)
        self.index += 1

    def allocate_memory_2(self, size: int) -> None:
        self.memory[self.index] = bytearray(3 * size)
        self.index += 1


if __name__ == "__main__":
    # Simple application that creates two threads with different durations:
    # - MainThread runs target() for 2 seconds
    # - Worker Thread-1 runs target() for 1 second
    # The profiler should capture both threads with their respective durations.
    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    Target().run()

    prof.stop()
