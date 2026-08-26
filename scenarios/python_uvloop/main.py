"""uvloop gate workload: asyncio tasks with CPU-bound and I/O-bound phases.

Wall-time expectations (profile.json) are burn-in anchors, not naive ratios of
the duration constants below. asyncio runs on a single thread: cpu_bound_work
busy-loops without awaiting, so tasks run sequentially when scheduled, not in
parallel on a shared wall-clock timeline. Samples also attribute nested calls
(cpu_bound_work inside mixed_workload) to the innermost matching stack frame.
"""

import asyncio
import os
import time

import uvloop
from ddtrace.profiling import Profiler


async def cpu_bound_work(duration: float) -> None:
    end_time: float = time.monotonic() + duration
    count: int = 0
    while time.monotonic() < end_time:
        count += 1


async def io_simulation(duration: float) -> None:
    await asyncio.sleep(duration)


async def mixed_workload(cpu_duration: float, io_duration: float) -> None:
    await cpu_bound_work(cpu_duration)
    await io_simulation(io_duration)


async def main() -> None:
    # Durations are fractions of EXECUTION_TIME_SEC (default 5s). Tasks are
    # created upfront; main runs its own cpu_bound_work before gather, so work
    # is interleaved on one thread rather than overlapping like parallel threads.
    execution_time_sec: int = int(os.environ.get("EXECUTION_TIME_SEC", "5"))

    tasks: list[asyncio.Task[None]] = [
        asyncio.create_task(cpu_bound_work(execution_time_sec * 0.3)),
        asyncio.create_task(mixed_workload(execution_time_sec * 0.2, execution_time_sec * 0.1)),
        asyncio.create_task(io_simulation(execution_time_sec * 0.4)),
    ]

    await cpu_bound_work(execution_time_sec * 0.3)
    await asyncio.gather(*tasks)


if __name__ == "__main__":
    prof: Profiler = Profiler()
    prof.start()

    uvloop.install()
    asyncio.run(main())
