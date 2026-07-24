import asyncio
import math

from ddtrace.profiling import Profiler

OFF_CPU_ITERATION_COUNT = 1_000
OFF_CPU_TASK_COUNT = 50
OFF_CPU_SLEEP_TIME = 0.001

ON_CPU_ITERATION_COUNT = 100
ON_CPU_TASK_COUNT = 1


async def off_cpu_task() -> None:
    for _ in range(OFF_CPU_ITERATION_COUNT):
        await asyncio.sleep(OFF_CPU_SLEEP_TIME)


async def on_cpu_task() -> None:
    for _ in range(ON_CPU_ITERATION_COUNT):
        math.factorial(3500)


async def main() -> None:
    await asyncio.gather(
        *(asyncio.create_task(off_cpu_task(), name=f"off_cpu_task") for _ in range(OFF_CPU_TASK_COUNT)),
        *(asyncio.create_task(on_cpu_task(), name=f"on_cpu_task") for _ in range(ON_CPU_TASK_COUNT)),
    )


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    asyncio.run(main())
