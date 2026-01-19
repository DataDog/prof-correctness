import asyncio


async def async_main() -> None:
    total_tasks = 0
    usages = []
    while True:
        print("Start iteration", flush=True)
        tasks = [asyncio.gather(*[asyncio.sleep(0.1) for _ in range(10)]) for _ in range(10)]
        # tasks = [asyncio.sleep(1) for _ in range(10)]
        await asyncio.gather(*tasks)
        total_tasks += len(tasks * 10)
        print(f"Completed {total_tasks} tasks in total", flush=True)

        with open("/sys/fs/cgroup/memory.current", "r") as f:
            usage = int(f.read().strip())
            usages.append(usage)
            usages_avg = sum(usages[-10:]) / 10 if len(usages) > 10 else 0
            print(f"Usage: {usages_avg} kB", flush=True)


if __name__ == "__main__":
    print("Starting!", flush=True)
    asyncio.run(async_main())
