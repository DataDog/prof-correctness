import asyncio
import os
import time

import uvloop
from ddtrace.profiling import Profiler


async def cpu_bound_work(duration: float) -> None:
    end_time = time.monotonic() + duration
    count = 0
    while time.monotonic() < end_time:
        count += 1


async def io_simulation(duration: float) -> None:
    await asyncio.sleep(duration)


async def mixed_workload(cpu_duration: float, io_duration: float) -> None:
    await cpu_bound_work(cpu_duration)
    await io_simulation(io_duration)


async def main() -> None:
    execution_time_sec = int(os.environ.get("EXECUTION_TIME_SEC", "5"))

    tasks: list[asyncio.Task[None]] = [
        asyncio.create_task(cpu_bound_work(execution_time_sec * 0.3)),
        asyncio.create_task(mixed_workload(execution_time_sec * 0.2, execution_time_sec * 0.1)),
        asyncio.create_task(io_simulation(execution_time_sec * 0.4)),
    ]

    await cpu_bound_work(execution_time_sec * 0.3)
    await asyncio.gather(*tasks)


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    uvloop.install()
    asyncio.run(main())
