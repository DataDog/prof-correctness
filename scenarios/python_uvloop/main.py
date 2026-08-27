"""uvloop gate: two named top-level tasks, sequential, no gather.

cpu_bound_work busy-loops for EXECUTION_TIME_SEC/2 (blocks the event loop).
io_simulation then sleeps for EXECUTION_TIME_SEC/2. Each is a named Task
awaited on its own so the main task is never a gather parent and never a
leaf doing work. Wall-time closed form: 50% cpu_task / 50% io_task.
"""

from __future__ import annotations

import asyncio
import os
import time

import uvloop
from ddtrace.profiling import Profiler


async def cpu_bound_work(duration: float) -> None:
    end_time: float = time.monotonic() + duration
    while time.monotonic() < end_time:
        pass


async def io_simulation(duration: float) -> None:
    await asyncio.sleep(duration)


async def main() -> None:
    execution_time_raw: str = os.environ.get("EXECUTION_TIME_SEC", "5")
    half: float = float(execution_time_raw) / 2.0
    await asyncio.create_task(cpu_bound_work(half), name="cpu_task")
    await asyncio.create_task(io_simulation(half), name="io_task")


if __name__ == "__main__":
    prof: Profiler = Profiler()
    prof.start()
    uvloop.install()
    asyncio.run(main())
