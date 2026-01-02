import asyncio
import os

from ddtrace.profiling import Profiler


async def my_coroutine(n: float) -> None:
    await asyncio.sleep(n)


async def main() -> None:
    # Simple application that creates two Tasks with different durations:
    # - "unnamed Task" runs my_coroutine() for 2 second
    # - short_task runs my_coroutine() for 1 second
    # The profiler should capture both Tasks with their respective durations.

    # Note: there currently is an issue in ddtrace that attaches one of the gathered Tasks to the "parent"
    # task. As a result, using an explicit name for the manually started Task would result in flakiness (as we cannot
    # know which Task name will be "absorbed" by the Parent).
    # For the time being, we thus don't name the Task, so that we will always have Task-1 and Task-2 in the Profile.

    # Note: additionally, there is an issue in how we count wall time that results in blatantly incorrect results.
    # We are in the process of making asyncio better in dd-trace-py; we will update the correctness check once that
    # issue is fixed.

    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    # Give the Profiler some time to start up
    await asyncio.sleep(0.5)

    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "2"))

    short_task = asyncio.create_task(my_coroutine(execution_time_sec / 2))

    # asyncio.gather will automatically wrap my_coroutine into a Task
    await asyncio.gather(short_task, my_coroutine(execution_time_sec))


if __name__ == "__main__":
    asyncio.run(main())
