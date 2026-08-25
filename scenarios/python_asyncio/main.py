import asyncio
import os


async def my_coroutine(n: float) -> None:
    await asyncio.sleep(n)


async def main() -> None:
    # Give the Profiler some time to start up
    await asyncio.sleep(0.5)

    execution_time_sec: float = float(os.environ.get("EXECUTION_TIME_SEC", "5"))

    # Two Tasks with different durations (EXECUTION_TIME_SEC default 5s in Dockerfiles):
    # - unnamed Task runs my_coroutine() for execution_time_sec
    # - short_task runs my_coroutine() for execution_time_sec / 2
    # The profiler should capture both Tasks with their respective durations.

    short_task: asyncio.Task[None] = asyncio.create_task(my_coroutine(execution_time_sec / 2.0), name="short_task")

    # asyncio.gather will automatically wrap my_coroutine into a Task
    await asyncio.gather(short_task, my_coroutine(execution_time_sec))


if __name__ == "__main__":
    asyncio.run(main())
