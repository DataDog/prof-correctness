# python_thread_churn_3.11

Verifies that work done in short-lived threads is still attributed correctly
under constant thread creation and destruction. All other scenarios use
long-lived threads, so the thread registration/unregistration hooks and the
sampler's tolerance for threads that vanish mid-sample are otherwise untested.

The workload repeatedly creates one `churn` thread that does a fixed chunk of
CPU work in `churn_worker`, joins it, and immediately spawns the next, for the
full duration (hundreds of create/destroy cycles).

## Expected behavior

- **cpu-time**: the large majority of CPU is attributed to `churn_worker`,
  demonstrating that ephemeral-thread work is captured despite the registration
  churn.
