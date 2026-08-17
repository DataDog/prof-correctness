# ddprof_julia

The test case was aimed at checking languages where we have fibers. 
When allocation profiling is active, we can be on a stack different form the thread.

This was causing crashes in ddprof
https://github.com/DataDog/ddprof/pull/213

Symbols are also interesting in Julia. Symbols are published in a .debug folder.
Test case should be adapted once these are processed.