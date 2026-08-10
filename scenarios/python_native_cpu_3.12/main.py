import binascii
import hashlib
import math
import os
import re
import zlib
from time import time

from ddtrace.profiling import Profiler

# 1 MB of compressible data shared across workers to avoid allocation overhead.
_DATA: bytes = b"abcdefgh" * (128 * 1024)
_TEXT: str = "the quick brown fox jumps over the lazy dog " * 5000
_PATTERN: re.Pattern[str] = re.compile(r"\b\w+\b")


def hash_work(duration: float) -> None:
    """Burn CPU in hashlib.sha256 (C extension)."""
    end = time() + duration
    while time() < end:
        hashlib.sha256(_DATA).digest()


def compress_work(duration: float) -> None:
    """Burn CPU in zlib.compress (C extension)."""
    end = time() + duration
    while time() < end:
        zlib.compress(_DATA, 6)


def factorial_work(duration: float) -> None:
    """Burn CPU in math.factorial (C extension)."""
    end = time() + duration
    while time() < end:
        math.factorial(100_000)


def regex_work(duration: float) -> None:
    """Burn CPU in re.findall (C extension regex engine)."""
    end = time() + duration
    while time() < end:
        _PATTERN.findall(_TEXT)


def crc_work(duration: float) -> None:
    """Burn CPU in binascii.crc32 (C extension)."""
    end = time() + duration
    while time() < end:
        binascii.crc32(_DATA)


if __name__ == "__main__":
    prof = Profiler()
    prof.start()

    execution_time = float(os.environ.get("EXECUTION_TIME_SEC", "25"))
    all_functions = [hash_work, compress_work, factorial_work, regex_work, crc_work]

    runs = 3
    time_per_function = execution_time / len(all_functions)
    time_per_run = time_per_function / runs

    for _ in range(runs):
        for func in all_functions:
            func(time_per_run)
