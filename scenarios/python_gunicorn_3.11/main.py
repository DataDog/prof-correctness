"""
This is a correctness check for ddtrace with a very basic (sync) Gunicorn application.
Gunicorn will run in the process we start and create one worker process.

What we do is start a thread that will issue HTTP requests to the Gunicorn application.
Processing this HTTP request will consume some CPU time (that we should profile).
After a while, this threads sends a special HTTP request, asking to stop the application.
"""

import os
import signal
import time
from collections.abc import Callable, Iterator
from threading import Thread

import requests


def make_request() -> None:
    expected_duration = int(os.environ.get("EXECUTION_TIME_SEC", "5"))

    start = time.monotonic()
    while time.monotonic() - start < expected_duration:
        try:
            requests.get("http://localhost:8000", timeout=1)
        except Exception:  # noqa: PERF203,BLE001
            print("Error making request")

    requests.get("http://localhost:8000/stop", timeout=1)


def sub_process(x: float) -> float:
    return x * 1.0001


def process(target: int) -> tuple[float, int]:
    x = 1.0
    i = 0
    while x < target:
        x = sub_process(x)
        i += 1

    return x, i


def app(environ: dict[str, str], start_response: Callable[[str, list[tuple[str, str]]], None]) -> Iterator[bytes]:
    x, iterations = process(target=123)
    data = f"Result: {(x, iterations)}".encode()
    status = "200 OK"
    response_headers = [("Content-type", "text/plain"), ("Content-Length", str(len(data)))]

    if environ["RAW_URI"] == "/stop":
        os.kill(os.getppid(), signal.SIGQUIT)

    start_response(status, response_headers)
    return iter([data])


t = Thread(target=make_request, name="Requester")
t.daemon = True
t.start()
