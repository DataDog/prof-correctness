Goal of this test case is to check the nodejs wall/cpu profiler.
3 functions (a, b and c) are run sequentially in a loop until requested execution time is reached (typically 10s).
`a` does a fixed number of arithmetic operations, `b` does double the number of operations compared to `a`, and `c` does asynchronous cryptographic operations during the same time as `b`.
At the start of the process a worker thread is spawned that execute the same work as the main thread but for half the execution time (ie. 5s).

Asynchronous cryptographic operations are run on separate libuv threads, and therefore are not seen in the samples captured by the wall/cpu profiler (since the samples are taken only on the javascript threads).
CPU consumed on non-JS threads is only reported in the profile of the main JS thread with a special `(non-JS threads)` frames because we are currenlty unable to correctly assign it to worker/main thread.

In the end we expect the following profiles:
* Main thread (10s execution): 
    * `a`: ~2.1s of wall/cpu
    * `b`: ~4.2s of wall/cpu
    * `(non-JS threads)`: ~6.9s of cpu (main crypto work + worker crypto work + system overhead)
* Worker thread (5s execution):
    * `a`: ~1.1s of wall/cpu
    * `b`: ~2.1s of wall/cpu

Note: Actual values may vary by ±10-15% due to system scheduling, startup overhead, and timer precision. 
Locally the accuracy is good, though we get 10-15% in CI.