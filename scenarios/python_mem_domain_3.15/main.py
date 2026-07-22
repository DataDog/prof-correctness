import os
import time

from ddtrace.profiling import Profiler

# Keep MEM-domain allocations live across heap snapshot uploads.
LIVE: list[bytearray] = []

BUF_BYTES = 16 * 1024 * 1024


def allocate_mem_domain_buffer() -> None:
    LIVE.append(bytearray(BUF_BYTES))


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    allocate_mem_domain_buffer()
    time.sleep(int(os.environ.get("EXECUTION_TIME_SEC", "15")))

    prof.stop()
