import binascii
import hashlib
import math
import os
import re
import zlib
from collections.abc import Callable
from time import time
from typing import Final

from ddtrace.profiling import Profiler

# 1 MB of compressible data shared across workers to avoid allocation overhead.
_DATA: Final[bytes] = b"abcdefgh" * (128 * 1024)
_TEXT: Final[str] = "the quick brown fox jumps over the lazy dog " * 5000
_PATTERN: Final[re.Pattern[str]] = re.compile(r"\b\w+\b")


def hash_work(duration: float) -> None:
    """Burn CPU in hashlib.sha256 (C extension)."""
    end: float = time() + duration
    while time() < end:
        hashlib.sha256(_DATA).digest()


def compress_work(duration: float) -> None:
    """Burn CPU in zlib.compress (C extension)."""
    end: float = time() + duration
    while time() < end:
        zlib.compress(_DATA, 6)


def factorial_work(duration: float) -> None:
    """Burn CPU in math.factorial (C extension)."""
    end: float = time() + duration
    while time() < end:
        math.factorial(100_000)


def regex_work(duration: float) -> None:
    """Burn CPU in re.findall (C extension regex engine)."""
    end: float = time() + duration
    while time() < end:
        _PATTERN.findall(_TEXT)


def crc_work(duration: float) -> None:
    """Burn CPU in binascii.crc32 (C extension)."""
    end: float = time() + duration
    while time() < end:
        binascii.crc32(_DATA)


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()

    execution_time: float = float(os.environ.get("EXECUTION_TIME_SEC", "25"))
    all_functions: list[Callable[[float], None]] = [
        hash_work,
        compress_work,
        factorial_work,
        regex_work,
        crc_work,
    ]

    runs: int = 3
    time_per_function: float = execution_time / len(all_functions)
    time_per_run: float = time_per_function / runs

    for _ in range(runs):
        for func in all_functions:
            func(time_per_run)


if __name__ == "__main__":
    main()
