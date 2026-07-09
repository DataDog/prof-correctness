# Node.js Allocation Profiling Test

This test validates Node.js allocation profiling by reusing the same controlled
allocation pattern as `node_heap`, with `DD_PROFILING_ALLOCATION_ENABLED=1` and
`@datadog/pprof` pinned to `5.15.1`.

The scenario runs on Node.js 26 because Node allocation profiling is only
available there. The `@datadog/pprof` package must therefore provide a native
prebuild for Node ABI 147.

## Test Behavior

The test creates two allocation functions:
- `a(size, refs)`: Allocates strings of 2MB each
- `b(size, refs)`: Allocates strings of 4MB each (2x larger than `a`)

Both functions are called once per iteration in a timed loop, creating equal numbers of allocations but different total allocated bytes.

## Expected Profiling Results

The allocation-mode heap profile includes both live (`inuse_*`) and total
allocated (`alloc_*`) sample types.

**In-use Objects Profile**: `inuse_objects`
- Expected: `0` for `a` and `b` `slice` stacks, because the test clears those references before the profile is exported

**Allocated Objects Profile**: `alloc_objects`
- Expected: `100` for `a` and `100` for `b`, within 5% error margin

**In-use Space Profile**: `inuse_space`
- Expected: `0` for `a` and `b` `slice` stacks, for the same reason as `inuse_objects`

**Allocated Space Profile**: `alloc_space`
- Expected: `33%` `a` / `66%` `b`, proportional to allocation sizes, within 5% error margin
