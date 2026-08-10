# FastAPI correctness check

Correctness check for Profiling with FastAPI + Uvicorn (asyncio).

## What it does

- Starts a FastAPI app on port 8000
- A background thread issues 30 concurrent HTTP requests continuously, in a loop
- Each request performs CPU-bound work (`process` → `sub_process`) and async I/O (`io_bound_work`)
- After `EXECUTION_TIME_SEC` (default 5s), the app shuts down

## Expected profile

The profile should show CPU time in:
- `process` / `sub_process` — the CPU-bound loop
- `io_bound_work` — async sleep (should show minimal CPU)
- `_WorkItem.run;_do_request` — the HTTP request handler (should show minimal CPU)
- `make_requests` — the background thread that issues the HTTP requests (should show minimal CPU)
