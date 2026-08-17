# Description

A simple test that allocates/frees memory and periodically leaks (no free) memory.

## Why is it not 100% of the inuse-space ?

Although the leak is the only "user" in-use memory, there are other allocations associated to the use of C++ (and exceptions).
Depending on load order, these allocations will be visible.
