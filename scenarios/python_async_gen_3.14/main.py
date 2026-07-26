import asyncio
import os


async def async_gen_work() -> None:
    async def ticker() -> object:
        for i in range(1_000_000):
            yield i
            if i % 10_000 == 0:
                await asyncio.sleep(0)

    async for _ in ticker():
        pass


async def main() -> None:
    await asyncio.sleep(0.5)
    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "8"))
    await asyncio.wait_for(async_gen_work(), timeout=execution_time_sec)


if __name__ == "__main__":
    asyncio.run(main())
