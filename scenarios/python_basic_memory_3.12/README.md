## Basic Memory Profiling

This test validates that the Datadog Python profiler correctly profiles memory allocations.

## Test Application

- Creates a list of `None` of a target size (1e6 items)
- The list is then filled by two functions `allocate_memory_1` and `allocate_memory_2` that allocate
  1024 bytes and 3 \* 1024 bytes respectively. The list is filled until the last item is not `None`.

## Expected Profile

Frames use qualified names (`Target.run`). Memory samples carry the numeric thread ID as `thread name`,
so the label check only requires a number.

### Samples

Those represent the number of times the function called the allocator. The values below are what
dd-trace-py `main` reports today, not the exact counts (33% / 44% / 22%). See https://github.com/DataDog/prof-correctness/pull/230.

- `^<module>;Target\.run;Target\.allocate_memory_1$`: 44% +/- 5% of the samples.
- `^<module>;Target\.run;Target\.allocate_memory_2$`: 39% +/- 5% of the samples.
- `^<module>;Target\.__init__;Target\.grow_list$`: 16% +/- 5% of the samples.

### Space

- `^<module>;Target\.run;Target\.allocate_memory_1$`: 25% +/- 5% of the space (1024 bytes)
- `^<module>;Target\.run;Target\.allocate_memory_2$`: 71% +/- 5% of the space (3 \* 1024 bytes)
- `^<module>;Target\.__init__;Target\.grow_list$`: 5% +/- 3% of the space (ints and the list buffer)
