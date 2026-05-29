## Basic Memory Profiling

This test validates that the Datadog Python profiler correctly profiles memory allocations.

## Test Application

- Creates a list of `None` of a target size (1e6 items)
- The list is then filled by two functions `allocate_memory_1` and `allocate_memory_2` that allocate
  1024 bytes and 3 \* 1024 bytes respectively. The list is filled until the last item is not `None`.

## Expected Profile

### Samples

Those represent the number of times the function called the allocator. This is hard to estimate manually
because of how the sampling logic works
(see [here](https://github.com/datadog/dd-trace-py/blob/9af78604497d1993826c59c351e1a4e53f817783/ddtrace/profiling/collector/_memalloc_tb.cpp#L329-L360)),
but the numbers are stable.

- `^<module>;run;allocate_memory_1$` should be present and should account for 25% of the samples.
- `^<module>;run;allocate_memory_2$` should be present and should account for 45% of the samples.
- `^<module>;__init__;grow_list$` should be present and should account for 20% of the samples.

### Space

- `^<module>;run;allocate_memory_1$` should be present and should account for 25% of the space (1024 bytes)
- `^<module>;run;allocate_memory_2$` should be present and should account for 75% of the space (3 \* 1024 bytes)
- `^<module>;__init__;grow_list$` should be present and should account for 3% of the space
  (list container, noise compared to the other allocations)
