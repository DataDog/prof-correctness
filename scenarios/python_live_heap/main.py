from __future__ import annotations

import os
import time
from typing import Final

from ddtrace.profiling import Profiler

# Module-scope retention keeps objects live across heap snapshot uploads.
LIVE: list[Target] = []

OBJ_SIZE: Final[int] = 16384  # matches DD_PROFILING_HEAP_SAMPLE_SIZE
N_MAJOR: Final[int] = 1600  # ~80% of live set
N_MINOR: Final[int] = 400  # ~20%


class Target:
    def __init__(self) -> None:
        self.live: list[bytes] = []

    def run(self, hold_seconds: float) -> None:
        self.retain_major()
        self.retain_minor()
        deadline: float = time.monotonic() + hold_seconds
        now: float = time.monotonic()
        while now < deadline:
            time.sleep(0.5)
            now = time.monotonic()

    def retain_major(self) -> None:
        for _ in range(N_MAJOR):
            chunk: bytes = bytes(OBJ_SIZE)
            self.live.append(chunk)

    def retain_minor(self) -> None:
        for _ in range(N_MINOR):
            chunk: bytes = bytes(OBJ_SIZE)
            self.live.append(chunk)


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()

    execution_time_raw: str = os.environ.get("EXECUTION_TIME_SEC", "15")
    execution_time: int = int(execution_time_raw)
    target: Target = Target()
    LIVE.append(target)
    target.run(hold_seconds=execution_time)

    prof.stop()


if __name__ == "__main__":
    main()
