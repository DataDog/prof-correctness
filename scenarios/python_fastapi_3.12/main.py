"""
Correctness check for ddtrace with FastAPI + Uvicorn (asyncio).

A background thread issues HTTP requests to the FastAPI application.
Processing these requests consumes CPU time (that we should profile).
After EXECUTION_TIME_SEC, the thread sends a shutdown request.
"""

import asyncio
import contextlib
import math
import os
import random
import signal
import time
from concurrent.futures import ThreadPoolExecutor
from threading import Event, Thread

import httpx
from fastapi import FastAPI, Request

app = FastAPI()

stop_event = Event()


def _do_request(client: httpx.Client) -> None:
    """Make a single request, repeat until stop_event is set."""
    while not stop_event.is_set():
        with contextlib.suppress(Exception):
            client.get("http://localhost:8000/", timeout=5)


def make_requests() -> None:
    """Maintain 20 concurrent requests for the configured duration."""
    expected_duration = int(os.environ.get("EXECUTION_TIME_SEC", "10"))

    n_requests = 30

    with httpx.Client() as client:
        with ThreadPoolExecutor(max_workers=n_requests) as executor:
            # Start 20 workers that continuously make requests
            futures = [executor.submit(_do_request, client) for _ in range(n_requests)]

            # Wait for the expected duration
            time.sleep(expected_duration)

            # Signal all workers to stop
            stop_event.set()

            # Wait for all workers to finish
            for f in futures:
                f.result()

        client.get("http://localhost:8000/stop", timeout=1)


def sub_process(x: float) -> float:
    """CPU-bound sub-operation."""
    for _ in range(100):
        x = x * math.pow(x, x)

    return x


async def process() -> tuple[float, int]:
    """CPU-bound async work that should appear in profiles."""
    x = 1.0
    i = 0
    start = time.monotonic()
    while time.monotonic() - start < 0.2:
        x = sub_process(x)
        i += 1

    return x, i


async def io_bound_work() -> None:
    await asyncio.sleep(0.3)
    if random.random() < 0.2:  # noqa: S311
        await io_bound_work()


@app.get("/")
async def root(request: Request) -> dict[str, str | int | float]:
    """Handle requests with CPU-bound work."""
    print(
        f"Processing request: "
        f"{request.client.host if request.client else 'unknown'}:{request.client.port if request.client else 'unknown'}"
    )
    x, iterations = await process()
    await io_bound_work()
    return {"result": x, "iterations": iterations}


@app.get("/stop")
async def stop() -> dict[str, str]:
    """Shutdown the application."""
    os.kill(os.getpid(), signal.SIGINT)
    return {"status": "stopping"}


@app.on_event("startup")
async def startup_event() -> None:
    """Start the requester thread on application startup."""
    t = Thread(target=make_requests, name="Requester")
    t.daemon = True
    t.start()
