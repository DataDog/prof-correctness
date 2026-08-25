import asyncio
import os


async def my_coroutine(n: float) -> None:
    await asyncio.sleep(n)


async def main() -> None:
    # Simple application that creates two Tasks with different durations:
    # - "unnamed Task" runs my_coroutine() for 2 second
    # - short_task runs my_coroutine() for 1 second
    # The profiler should capture both Tasks with their respective durations.

    # Give the Profiler some time to start up
    await asyncio.sleep(0.5)

    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "5"))

    short_task = asyncio.create_task(my_coroutine(execution_time_sec / 2.0), name="short_task")

    # asyncio.gather will automatically wrap my_coroutine into a Task
    await asyncio.gather(short_task, my_coroutine(execution_time_sec))


if __name__ == "__main__":
    asyncio.run(main())
