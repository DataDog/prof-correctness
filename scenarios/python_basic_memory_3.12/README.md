## Basic Memory Profiling

This test validates that the Datadog Python profiler correctly profiles memory allocations.

## Test Application

- Creates a list of `None` of a target size (1e6 items)
- The list is then filled by two functions `allocate_memory_1` and `allocate_memory_2` that allocate
  1024 bytes and 3 \* 1024 bytes respectively. The list is filled until the last item is not `None`.

## Expected Profile

The expected shares are the exact allocation counts and bytes of this workload, not values copied
from past profiles. They were measured on CPython 3.12 with a small C allocator hook that counts every
`PYMEM_DOMAIN_OBJ` and `PYMEM_DOMAIN_MEM` call (malloc, calloc and realloc), without the profiler:

| Function | Domain | Allocations | Bytes |
|---|---|---|---|
| `allocate_memory_1` | obj | 1,499,872 (33.3%) | 556 MB (24.6%) |
| `allocate_memory_2` | obj | 1,999,872 (44.4%) | 1,596 MB (70.6%) |
| `grow_list` | obj | 999,747 (22.2%) | 32 MB (1.4%) |
| `grow_list` | mem | 86 (0.0%) | 76 MB (3.4%) |

- `allocate_memory_1` makes 3 allocations per call: the `bytearray` object, its 1024-byte buffer,
  and the int for `self.index + 1`.
- `allocate_memory_2` makes 4: the same 3, plus the int for `3 * size`.
- `grow_list` allocates one int per `range` item (obj) and grows the list buffer with
  `PyMem_Realloc` (mem).

Each expectation matches the `allocator domain` label (`obj` or `mem`). Memory samples carry the
numeric thread ID as `thread name` (dd-trace-py 4.5.0 and later read it without calling Python code),
so the check only requires a numeric value. Frames use qualified names such as `Target.run`.

### Samples (`alloc-samples`)

- `allocate_memory_1`, obj: 33% +/- 5%
- `allocate_memory_2`, obj: 44% +/- 5%
- `grow_list`, obj: 22% +/- 5%

### Space (`alloc-space`)

- `allocate_memory_1`, obj: 25% +/- 5%
- `allocate_memory_2`, obj: 71% +/- 5%
- `grow_list`, obj: 1% +/- 1%
- `grow_list`, mem: 3% +/- 2%
