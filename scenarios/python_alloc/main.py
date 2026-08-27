"""Allocation workload with two distinct sites for profile attribution.

Each loop iteration calls allocate_memory_1 (size bytes) and allocate_memory_2
(3x size bytes), producing a 1:3 alloc-space ratio in profiles.

Objects are not retained: this gate asserts allocation events, not the live set
(python_live_heap covers persistent heap). Holding 1M bytearrays (~2 GB RSS)
would also couple the run to TRACEBACK_ARRAY_MAX_COUNT (65535 live samples).
"""

from __future__ import annotations

from typing import Final

from ddtrace.profiling import Profiler

# run() performs pair count = _CAPACITY // 2 iterations (500000 each site).
_CAPACITY: Final[int] = int(1e6)
_ALLOC_SIZE: Final[int] = 1024


class Target:
    def run(self) -> None:
        for _ in range(_CAPACITY // 2):
            self.allocate_memory_1(_ALLOC_SIZE)
            self.allocate_memory_2(_ALLOC_SIZE)

    def allocate_memory_1(self, size: int) -> None:
        bytearray(size)

    def allocate_memory_2(self, size: int) -> None:
        bytearray(3 * size)


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()

    target: Target = Target()
    target.run()

    prof.stop()


if __name__ == "__main__":
    main()
