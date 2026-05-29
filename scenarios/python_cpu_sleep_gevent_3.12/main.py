from gevent import monkey

monkey.patch_all()

import os  # noqa: E402
import time  # noqa: E402

import gevent  # noqa: E402


def cpu_burst(target_seconds: float) -> int:
    deadline = time.monotonic() + target_seconds
    x = 0x1234
    while time.monotonic() < deadline:
        x = (x * 48271) % 2147483647
    return x


def slot(cpu_seconds: float, off_cpu: float, deadline: float) -> None:
    while time.monotonic() < deadline:
        time.sleep(off_cpu)
        cpu_burst(cpu_seconds)


def main() -> None:
    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    cpu_seconds = 0.01
    off_cpu = 0.05
    concurrency = 4
    stagger = 0.01

    deadline = time.monotonic() + execution_time_sec
    greenlets = []
    for i in range(concurrency):
        if i > 0:
            gevent.sleep(stagger)
        greenlets.append(gevent.spawn(slot, cpu_seconds, off_cpu, deadline))
    gevent.joinall(greenlets)


if __name__ == "__main__":
    main()
