import asyncio
import os
import time
from collections.abc import AsyncGenerator


async def async_gen_work(duration: float) -> None:
    async def ticker() -> AsyncGenerator[int, None]:
        end = time.monotonic() + duration
        i = 0
        while time.monotonic() < end:
            yield i
            i += 1
            if i % 1000 == 0:
                await asyncio.sleep(0)

    async for _ in ticker():
        pass


async def main() -> None:
    duration = float(os.environ.get("EXECUTION_TIME_SEC", "8"))
    await asyncio.create_task(async_gen_work(duration), name="gen_task")


if __name__ == "__main__":
    asyncio.run(main())
