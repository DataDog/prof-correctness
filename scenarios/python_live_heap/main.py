from __future__ import annotations

import os
import time

from ddtrace.profiling import Profiler

# Module-scope retention keeps objects live across heap snapshot uploads.
LIVE: list[Target] = []

OBJ_SIZE: int = 16384  # matches DD_PROFILING_HEAP_SAMPLE_SIZE
N_MAJOR: int = 1600  # ~80% of live set
N_MINOR: int = 400  # ~20%


class Target:
    live: list[bytes]

    def __init__(self) -> None:
        self.live = []

    def run(self, hold_seconds: float) -> None:
        self.retain_major()
        self.retain_minor()
        deadline: float = time.monotonic() + hold_seconds
        while time.monotonic() < deadline:
            time.sleep(0.5)

    def retain_major(self) -> None:
        for _ in range(N_MAJOR):
            self.live.append(bytes(OBJ_SIZE))

    def retain_minor(self) -> None:
        for _ in range(N_MINOR):
            self.live.append(bytes(OBJ_SIZE))


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()

    execution_time: int = int(os.environ.get("EXECUTION_TIME_SEC", "15"))
    target: Target = Target()
    LIVE.append(target)
    target.run(hold_seconds=execution_time)

    prof.stop()


if __name__ == "__main__":
    main()
