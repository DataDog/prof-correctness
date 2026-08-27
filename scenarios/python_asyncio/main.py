from __future__ import annotations

import asyncio
import os


async def my_coroutine(n: float) -> None:
    await asyncio.sleep(n)


async def main() -> None:
    execution_time_raw: str = os.environ.get("EXECUTION_TIME_SEC", "5")
    execution_time_sec: float = float(execution_time_raw)
    # 1:2 durations → 33% / 67% of wall-time on the two named tasks.
    short_task: asyncio.Task[None] = asyncio.create_task(my_coroutine(execution_time_sec / 2.0), name="short_task")
    long_task: asyncio.Task[None] = asyncio.create_task(my_coroutine(execution_time_sec), name="long_task")
    await asyncio.gather(short_task, long_task)


if __name__ == "__main__":
    asyncio.run(main())
