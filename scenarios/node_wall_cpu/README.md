Goal of this test case is to check the nodejs wall/cpu profiler.
3 functions (a, b and c) are run sequentially in a loop until requested execution time is reached (typically 10s).
`a` does a fixed number of arithmetic operations, `b` does double the number of operations compared to `a`, and `c` does asynchronous cryptographic operations during the same time as `b`.
At the start of the process a worker thread is spawned that execute the same work as the main thread but for half the execution time (ie. 5s).

Asynchronous cryptographic operations are run on separate libuv threads, and therefore are not seen in the samples captured by the wall/cpu profiler (since the samples are taken only on the javascript threads).
CPU consumed on non-JS threads is only reported in the profile of the main JS thread with a special `(non-JS threads)` frames because we are currenlty unable to correctly assign it to worker/main thread.

In the end we expect the following profiles:
* Main thread: 
    * `a`: 2s of wall/cpu
    * `b`: 4s of wall/cpu
    * `(non-JS threads)`: 4s (main) + 2s (worker) of cpu (+ 0.5s for some other work that occur outside of JS threads)
* Worker thread:
    * `a`: 1s of wall/cpu
    * `b`: 2s of wall/cpu 