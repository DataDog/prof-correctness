import os
import asyncio


async def my_coroutine(n) -> None:
    await asyncio.sleep(n)


async def main() -> None:
    from ddtrace.profiling import Profiler

    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    # Simple application that creates two Tasks with different durations:
    # - "unnamed Task" runs my_coroutine() for 2 second
    # - short_task runs my_coroutine() for 1 second
    # The profiler should capture both threads with their respective durations.

    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    short_task = asyncio.create_task(my_coroutine(EXECUTION_TIME_SEC / 2))

    # asyncio.gather will automatically wrap my_coroutine into a Task
    await asyncio.gather(short_task, my_coroutine(EXECUTION_TIME_SEC))


if __name__ == "__main__":
    asyncio.run(main())
