# python_native_cpu_3.11

Verifies that the Python profiler correctly attributes CPU time to Python frames that call native (C-extension) functions.

Five functions run sequentially, each for an equal share (~20%) of the total execution time. Each function spends its entire time in a tight loop calling a different CPU-intensive C extension:

| Function | Native call |
|---|---|
| `hash_work` | `hashlib.sha256` |
| `compress_work` | `zlib.compress` |
| `factorial_work` | `math.factorial` |
| `regex_work` | `re.Pattern.findall` |
| `crc_work` | `binascii.crc32` |

Adaptive sampling is disabled to ensure dense, near-fixed-rate sampling.

## Expected behavior

- **cpu-time**: each function appears at ~20% (±5%)
- **wall-time**: each function appears at ~20% (±5%)

A profiler that fails to count CPU time consumed inside C extensions (e.g. by only sampling at Python bytecode boundaries) would show all five functions with zero or near-zero cpu-time, causing this test to fail.
