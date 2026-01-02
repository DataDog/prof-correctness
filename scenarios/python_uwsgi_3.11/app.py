import os
import signal
import time
from collections.abc import Callable
from threading import Thread

import requests
import uwsgi  # pyright: ignore[reportMissingModuleSource]


def _make_requests() -> None:
    run_time = int(os.environ.get("EXECUTION_TIME_SEC", "10"))

    start = time.monotonic()
    while time.monotonic() - start < run_time:
        try:
            print(f"Requester PID: {os.getpid()} requesting")
            requests.get("http://localhost:8000", timeout=1)
        except Exception as e:  # noqa: PERF203,BLE001
            print(e)

    time.sleep(1.0)

    # Ask the master process to exit
    try:
        os.kill(uwsgi.masterpid(), signal.SIGINT)
    finally:
        pass


def compute_big_number() -> int:
    x = 0
    for i in range(1_000_000):
        x *= i
    return x


def application(environ: dict, start_response: Callable) -> list[bytes]:  # noqa: ARG001  # pyright: ignore[reportUnusedParameter]
    x = compute_big_number()
    start_response("200 OK", [("Content-Type", "text/plain")])
    return [f"Hello, World! {x}".encode()]


Thread(target=_make_requests, daemon=True, name="Requester").start()
