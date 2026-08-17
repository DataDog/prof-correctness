# Description

A simple test that allocates/frees memory and periodically leaks (no free) memory.

## Why is it not 100% of the inuse-space ?

Although the leak is the only "user" in-use memory, there are other allocations associated to the use of C++ (and exceptions).
Depending on load order, these allocations will be visible.

## alloc-space assertion

The test allocates via both `allocate_memory→operator new` and `leak_function→malloc`.
The alloc-space regex covers both paths. CI intermittently failed (~6% gap) when the
assertion only matched the `operator new` path and `leak_function` bytes dominated.
We fixed the regex rather than widening `error_margin`, since the flake was a coverage
gap rather than runtime variance.
