"""Test scenario for uvloop profiling."""

import asyncio
import os
import time

import uvloop
from ddtrace.profiling import Profiler


async def cpu_bound_work(duration: float) -> None:
    """Perform CPU-bound work for a specified duration."""
    end_time = time.monotonic() + duration
    count = 0
    while time.monotonic() < end_time:
        count += 1


async def io_simulation(duration: float) -> None:
    """Simulate I/O-bound work using asyncio.sleep."""
    await asyncio.sleep(duration)


async def mixed_workload(cpu_duration: float, io_duration: float) -> None:
    """Perform both CPU-bound and I/O-bound work."""
    await cpu_bound_work(cpu_duration)
    await io_simulation(io_duration)


async def main() -> None:
    """Main async function that runs the test workload."""
    execution_time_sec = int(os.environ.get("EXECUTION_TIME_SEC", "5"))

    # Run multiple concurrent tasks to exercise uvloop
    tasks = [
        asyncio.create_task(cpu_bound_work(execution_time_sec * 0.3)),
        asyncio.create_task(mixed_workload(execution_time_sec * 0.2, execution_time_sec * 0.1)),
        asyncio.create_task(io_simulation(execution_time_sec * 0.4)),
    ]

    # Also do some work in the main task
    await cpu_bound_work(execution_time_sec * 0.3)

    # Wait for all tasks to complete
    await asyncio.gather(*tasks)

    print(f"Completed execution for {execution_time_sec} seconds")


if __name__ == "__main__":
    # Start the profiler
    prof = Profiler()
    prof.start()

    # Install uvloop as the event loop
    uvloop.install()

    # Run the async main function
    asyncio.run(main())
