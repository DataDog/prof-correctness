from __future__ import annotations

import os
import time
from typing import Final

from ddtrace.profiling import Profiler

# Module-scope retention keeps the MEM-domain buffer live across the heap snapshot.
LIVE: list[Target] = []

BUFFER_SIZE: Final[int] = 16 * 1024 * 1024


class Target:
    def __init__(self) -> None:
        self.buffer: bytearray | None = None

    def run(self, hold_seconds: float) -> None:
        self.allocate_mem_domain_buffer()
        deadline: float = time.monotonic() + hold_seconds
        now: float = time.monotonic()
        while now < deadline:
            time.sleep(0.5)
            now = time.monotonic()

    def allocate_mem_domain_buffer(self) -> None:
        # bytearray's internal buffer is PYMEM_DOMAIN_MEM on 3.13+ (this gate is 3.14/3.15).
        self.buffer = bytearray(BUFFER_SIZE)


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
