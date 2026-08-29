import asyncio
import os
import time


def cpu_burst(target_seconds: float) -> int:
    deadline = time.monotonic() + target_seconds
    x = 0x1234
    while time.monotonic() < deadline:
        x = (x * 48271) % 2147483647
    return x


async def slot(cpu_seconds: float, off_cpu: float, deadline: float) -> None:
    while time.monotonic() < deadline:
        await asyncio.sleep(off_cpu)
        cpu_burst(cpu_seconds)


async def main() -> None:
    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    cpu_seconds = 0.01
    off_cpu = 0.05
    concurrency = 4
    stagger = 0.01

    deadline = time.monotonic() + execution_time_sec
    tasks: list[asyncio.Task[None]] = []
    for i in range(concurrency):
        if i > 0:
            await asyncio.sleep(stagger)
        tasks.append(asyncio.create_task(slot(cpu_seconds, off_cpu, deadline)))
    await asyncio.gather(*tasks)


if __name__ == "__main__":
    asyncio.run(main())
