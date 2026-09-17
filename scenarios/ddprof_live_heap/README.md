# Description

A simple test that allocates/frees memory and periodically leaks (no free) memory.

## Margin notes

- **inuse-space / inuse-objects (10% margin)**: Although the leak is the only "user" in-use memory, there are other allocations associated to the use of C++ (and exceptions). Depending on load order and CI runner characteristics, unaccounted allocations can reduce the leak_function visibility to ~91%.
- **alloc-space (5% margin)**: Widened regex covers both `operator new` and `leak_function` malloc paths, achieving consistent 100% coverage.
