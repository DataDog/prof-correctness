# python_safe_point_bias_3.12

Verifies that the Python profiler correctly attributes CPU time to `slow_method` rather than to `empty_method`, which is called from `slow_method` but does no work.

## What the program does

`slow_method` performs string concatenations and then calls `empty_method`, which is a no-op. The loop runs for the full `EXECUTION_TIME_SEC` duration.

```python
def empty_method() -> None:
    pass


def slow_method() -> None:
    while time() < end_time:
        x = "h" + "e" + "l" + "l" + "o" + ","
        x += "w" + "o" + "r" + "l" + "d"
        empty_method()
```

## Expected behavior

- **cpu-time**: `slow_method` appears in ~100% of samples (inclusive), since all CPU work happens there.
- **cpu-time**: `empty_method` appears in at most 10% of samples. Its call and
  return still consume some CPU, but the no-op should not receive a significant
  share of its caller's CPU time.

A profiler with a blame-attribution bug would incorrectly show `empty_method` consuming a substantial share of CPU time.
