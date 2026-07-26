import os
import time

from ddtrace.profiling import Profiler

# Allocations are held at module scope so they stay live for the whole process
# and therefore appear in every live-heap snapshot the profiler exports.
LIVE: list = []

# Two distinctly-named call sites that retain a known live set. Both allocate
# the SAME object size and differ only in count, so the 80/20 split holds for
# both metrics the live-heap profile reports:
#   - heap-space        (live bytes)  -> count * size  -> 80/20
#   - heap-live-samples (live objects) -> count        -> 80/20
# We use ``bytes`` (PyObject_Malloc / OBJ domain), which the heap profiler
# tracks identically across Python versions and independent of the MEM-domain
# toggle (``bytearray`` moved OBJ -> MEM in 3.13). This keeps the scenario a
# clean live-heap check across the 3.14 -> 3.15 migration.
OBJ_SIZE = 16384  # 16 KiB, well above the pymalloc small-object threshold
N_MAJOR = 1600  # ~80% of the live set  (1600 * 16 KiB ~= 25 MiB)
N_MINOR = 400  # ~20% of the live set  ( 400 * 16 KiB ~=  6 MiB)


class Target:
    def __init__(self) -> None:
        self.live: list = []

    def run(self, hold_seconds: float) -> None:
        self.retain_major()
        self.retain_minor()
        # Keep the live set alive across several upload intervals so each
        # exported heap snapshot contains the full set.
        deadline = time.monotonic() + hold_seconds
        while time.monotonic() < deadline:
            time.sleep(0.5)

    def retain_major(self) -> None:
        for _ in range(N_MAJOR):
            self.live.append(bytes(OBJ_SIZE))

    def retain_minor(self) -> None:
        for _ in range(N_MINOR):
            self.live.append(bytes(OBJ_SIZE))


if __name__ == "__main__":
    prof = Profiler()
    prof.start()  # As early as possible so the allocations below are sampled.

    execution_time = int(os.environ.get("EXECUTION_TIME_SEC", "15"))
    target = Target()
    LIVE.append(target)
    target.run(hold_seconds=execution_time)

    prof.stop()
