"""Allocation workload with two distinct sites for profile attribution.

Each loop iteration calls allocate_memory_1 (size bytes) and allocate_memory_2
(3x size bytes), producing a 1:3 alloc-space ratio in profiles.
"""

from ddtrace.profiling import Profiler

# run() fills two slots per iteration (pair count = len // 2).
_CAPACITY = int(1e6)
_ALLOC_SIZE = 1024


class Target:
    def __init__(self) -> None:
        self.memory: list[bytearray | None] = [None] * _CAPACITY
        self.index: int = 0

    def run(self) -> None:
        for _ in range(len(self.memory) // 2):
            self.allocate_memory_1(_ALLOC_SIZE)
            self.allocate_memory_2(_ALLOC_SIZE)

    def allocate_memory_1(self, size: int) -> None:
        self.memory[self.index] = bytearray(size)
        self.index += 1

    def allocate_memory_2(self, size: int) -> None:
        self.memory[self.index] = bytearray(3 * size)
        self.index += 1


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()

    target: Target = Target()
    target.run()

    prof.stop()


if __name__ == "__main__":
    main()
