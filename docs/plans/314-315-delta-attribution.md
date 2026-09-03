# 3.14 vs 3.15: ddtrace or CPython?

Parked compare deltas (same `da3df551` wheel, both interpreters):

- `python_live_heap` retain_major heap-live-samples: ~15pp

Question: last ~6 months of ddtrace, or CPython 3.14 vs 3.15?

## Method

The current gate already holds ddtrace fixed (`da3df551` / the #194 pin) and only varies the interpreter. A red compare on that pin is **not** “main drifted for 6 months.” It is interpreter, or this-wheel × interpreter.

- **ddtrace regression:** hold the interpreter fixed (3.14 and/or 3.15) and bisect ddtrace SHAs, or run the pair on an older S3 wheel from ~6 months ago vs the pin.
- **CPython:** same wheel, two interpreters — what the pair already does. If the delta survives on an old wheel that predates the suspected ddtrace commits, it is CPython.
