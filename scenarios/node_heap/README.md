# Node.js Heap Profiling Test

This test validates the Node.js heap profiler by creating a controlled allocation pattern and verifying the profiler correctly captures allocation sources.

## Test Behavior

The test creates two allocation functions:
- `a(size, refs)`: Allocates strings of 2MB each
- `b(size, refs)`: Allocates strings of 4MB each (2x larger than `a`)

Both functions are called once per iteration in a timed loop, creating equal **numbers** of allocations but different **sizes** of memory usage.

## Expected Profiling Results

**Objects Profile**: Should count allocation instances
- Theoretically: 50% `a` / 50% `b` (equal number of calls)
- Measured: Matches expectations

**Space Profile**: Should count memory usage  
- Expected: 33% `a` / 66% `b` (proportional to allocation sizes)
- Measured: Matches expectations
