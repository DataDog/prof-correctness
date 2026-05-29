import os
import time


def cpu_burst(target_seconds: float) -> int:
    deadline = time.monotonic() + target_seconds
    x = 0x1234
    while time.monotonic() < deadline:
        x = (x * 48271) % 2147483647
    return x


def main() -> None:
    execution_time_sec = float(os.environ.get("EXECUTION_TIME_SEC", "30"))
    cpu_seconds = 0.01
    off_cpu = 0.05
    deadline = time.monotonic() + execution_time_sec
    while time.monotonic() < deadline:
        time.sleep(off_cpu)
        cpu_burst(cpu_seconds)


if __name__ == "__main__":
    main()
