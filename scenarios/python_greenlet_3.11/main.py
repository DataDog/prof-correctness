"""Test scenario for greenlet profiling."""

import os
import time

import greenlet


class GreenletBurner:
    def __init__(self) -> None:
        self.greenlet_a = greenlet.greenlet(self.work_a)
        self.greenlet_b = greenlet.greenlet(self.work_b)

        self.total_count_a = 0
        self.total_count_b = 0

    def work_dependency(self, count: int) -> tuple[bool, int]:
        count += 1
        if count % 10000 == 0:
            return True, count

        return False, count

    def work_a(self) -> None:
        """Perform work in greenlet A."""
        end_time = time.monotonic() + 0.1
        count = 0
        while time.monotonic() < end_time:
            has_dependency, count = self.work_dependency(count)
            if has_dependency:
                self.greenlet_b.switch()

            self.total_count_a += 1

    def work_b(self) -> None:
        """Perform work in greenlet B."""
        end_time = time.monotonic() + 0.1
        count = 0
        while time.monotonic() < end_time:
            has_dependency, count = self.work_dependency(count)
            if has_dependency:
                self.greenlet_a.switch()

            self.total_count_b += 1

    def work(self) -> None:
        self.greenlet_a.switch()
        self.greenlet_b.switch()

        self.greenlet_a = greenlet.greenlet(self.work_a)
        self.greenlet_b = greenlet.greenlet(self.work_b)


def main() -> None:
    execution_time_sec = int(os.environ.get("EXECUTION_TIME_SEC", "10"))

    greenlet_burner = GreenletBurner()

    # Run for specified duration
    end_time = time.monotonic() + execution_time_sec

    while time.monotonic() < end_time:
        # Start switching between greenlets
        greenlet_burner.work()

    print(f"Completed execution for {execution_time_sec} seconds")


if __name__ == "__main__":
    main()
