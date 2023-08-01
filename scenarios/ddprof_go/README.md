# Description

The test aims at checking that we are able to load and capture C allocations.

# Shortcomings

- We are not unwinding through the ASM CGo frame.
- The quantity of allocations is hard to predict considering the Go allocator reserves mmap regions.
