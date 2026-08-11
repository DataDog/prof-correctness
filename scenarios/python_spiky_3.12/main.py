import math
import os
import random
from itertools import zip_longest
from time import sleep, time

from ddtrace.profiling import Profiler


def cpu_spike(duration: float) -> None:
    end = time() + duration
    while time() < end:
        math.factorial(50000)


def sleep_period(duration: float) -> None:
    sleep(duration)


def _generate_durations(total: float, min_dur: float, max_dur: float) -> list[float]:
    """Return durations drawn from [min_dur, max_dur] that sum exactly to total."""
    durations: list[float] = []
    remaining = total
    while remaining > min_dur:
        d = random.uniform(min_dur, min(max_dur, remaining))  # noqa: S311
        durations.append(d)
        remaining -= d
    if remaining > 0:
        durations.append(remaining)
    return durations


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "20"))
    # Split evenly so the cpu/wall-time percentages are a fixed, predictable 50/50.
    half = execution_time / 2.0

    # Precompute durations whose sums are exactly `half` each.  Individual
    # spike lengths are random but the totals are deterministic, so the
    # expected-profile percentages hold regardless of how the randomness plays out.
    cpu_durations = _generate_durations(half, 0.5, 1.5)
    sleep_durations = _generate_durations(half, 0.8, 1.2)

    for cpu_dur, sleep_dur in zip_longest(cpu_durations, sleep_durations):
        if cpu_dur is not None:
            cpu_spike(cpu_dur)
        if sleep_dur is not None:
            sleep_period(sleep_dur)
