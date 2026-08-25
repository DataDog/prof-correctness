"""Allocation workload with two distinct sites for profile attribution.

Each loop iteration calls allocate_memory_1 (size bytes) and allocate_memory_2
(3x size bytes), producing a 1:3 alloc-space ratio in profiles.
"""

from ddtrace.profiling import Profiler

# Capacity must be even: run() fills two slots per iteration.
_CAPACITY = int(1e6)
_ALLOC_SIZE = 1024


class Target:
    def __init__(self) -> None:
        self.memory: list[bytearray | None] = []
        self.index: int = 0
        self.grow_list(_CAPACITY)

    def run(self) -> None:
        while self.index < len(self.memory):
            self.allocate_memory_1(_ALLOC_SIZE)
            self.allocate_memory_2(_ALLOC_SIZE)

    def grow_list(self, target: int) -> None:
        self.memory = [None] * target

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
