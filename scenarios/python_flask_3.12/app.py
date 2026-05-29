import math
import os
import signal
import time
from threading import Thread

import requests
from flask import Flask

app = Flask(__name__)


def _make_requests() -> None:
    start = time.monotonic()
    while time.monotonic() - start < 5.0:
        try:
            requests.get("http://localhost:8000", timeout=1)
        except Exception as e:  # noqa: PERF203,BLE001
            print(f"Error making request: {e}")

    requests.get("http://localhost:8000/stop", timeout=1)


def compute_big_number() -> int:
    start = time.monotonic()
    x = 2
    while time.monotonic() - start < 0.5:
        x *= math.factorial(min(10_000, x))

    return min(x, 1e32)


@app.route("/")
def hello_world() -> str:
    x = compute_big_number()
    return f"<p>Hello, World! {x}</p>"


@app.route("/stop")
def stop() -> str:
    os.kill(os.getpid(), signal.SIGINT)
    return "Stopping..."


t = Thread(target=_make_requests, name="Requester")
t.daemon = True
t.start()
