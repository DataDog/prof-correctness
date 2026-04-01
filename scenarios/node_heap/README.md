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
- Actually measured: 33% `a` / 66% `b` (biased toward larger objects)

**Space Profile**: Should count memory usage  
- Expected: 33% `a` / 66% `b` (proportional to allocation sizes)
- Measured: Matches expectations

## Open Questions

The objects profile shows bias toward larger allocations (66% vs 50% expected), suggesting the Node.js heap profiler may be:
- Sampling based on allocation size rather than allocation count
- Using size-weighted sampling for performance reasons
- Exhibiting V8 engine optimization effects on string allocation

**TODO**: Clarify with the Node.js profiler team how sampling works and why there's bias toward larger objects in the objects profile.
