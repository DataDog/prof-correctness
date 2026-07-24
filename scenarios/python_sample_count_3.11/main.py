import asyncio
import time
import math

from ddtrace.profiling import Profiler


SLEEP_TIME = 0.0001
SLEEP_COUNT = 1_000
ON_CPU_COUNT = 10
OFF_CPU_COUNT = 500


async def off_cpu_task() -> None:
    for _ in range(SLEEP_COUNT):
        await asyncio.sleep(SLEEP_TIME)


async def on_cpu_task() -> None:
    for x in range(ON_CPU_COUNT):
        math.factorial(x)


async def main() -> None:
    await asyncio.gather(*(off_cpu_task() for _ in range(OFF_CPU_COUNT)), *(on_cpu_task() for _ in range(ON_CPU_COUNT)))


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    start = time.time()
    asyncio.run(main())
    end = time.time()
    print(f"Run took {end - start} seconds")
