import os
import signal
import time
from threading import Thread

import requests
import uwsgi


def _make_requests() -> None:
    start = time.monotonic()
    while time.monotonic() - start < 5.0:
        try:
            requests.get("http://localhost:8000", timeout=1)
        except Exception as e:  # noqa: PERF203,BLE001
            print(e)

    # Ask the master process to exit
    os.kill(uwsgi.masterpid(), signal.SIGINT)


def compute_big_number() -> int:
    x = 0
    for i in range(1_000_000):
        x += i
    return x


def application(environ: dict, start_response: callable) -> list[bytes]:  # noqa: ARG001
    x = compute_big_number()
    start_response("200 OK", [("Content-Type", "text/plain")])
    return [f"Hello, World! {x}".encode()]


Thread(target=_make_requests, daemon=True).start()
