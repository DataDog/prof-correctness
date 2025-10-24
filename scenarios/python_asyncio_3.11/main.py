import os
import asyncio
from ddtrace.profiling import Profiler


async def my_coroutine(n):
    await asyncio.sleep(n)


async def main():
    # Simple application that creates two threads with different durations:
    # - MainThread runs target() for 2 seconds
    # - Worker Thread-1 runs target() for 1 second
    # The profiler should capture both threads with their respective durations.
    prof = Profiler()
    prof.start()  # Should be as early as possible, eg before other imports, to ensure everything is profiled

    EXECUTION_TIME_SEC = int(os.environ.get("EXECUTION_TIME_SEC", "2"))

    t = asyncio.create_task(my_coroutine(EXECUTION_TIME_SEC / 2))
    await asyncio.gather(t, my_coroutine(EXECUTION_TIME_SEC))


if __name__ == "__main__":
    asyncio.run(main())
