import asyncio
import math
import os
import time


async def my_coroutine(n: float) -> None:
    await asyncio.sleep(n)


async def long_computation(seconds: float) -> int:
    start_time = time.monotonic()
    result = 1
    i = 0
    while time.monotonic() < start_time + seconds:
        result *= math.factorial(10)

        # Yield control back every once in a while, important for all Tasks to appear
        i += 1
        if i % 1000 == 0:
            await asyncio.sleep(0.0)

    return result


async def async_main() -> None:
    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "3"))

    # Give the Profiler some time to start up; this we don't check in the expected Profile.
    await asyncio.sleep(0.5)

    # This is going to run in the background, the stack should not include async_main
    short_task = asyncio.create_task(my_coroutine(execution_time_sec / 2), name="short_task")

    # Yield control back so that short_task can actually start
    await asyncio.sleep(0.0)

    # Do on-CPU work for a while
    await long_computation(execution_time_sec / 3.0)

    # asyncio.gather will automatically wrap my_coroutine into a Task
    # Here we are explicitly "linking" the current Task and short_task / the new Task, so
    # we expect to see "async_main" in their Stacks.
    await asyncio.gather(short_task, my_coroutine(execution_time_sec))


if __name__ == "__main__":
    asyncio.run(async_main())
