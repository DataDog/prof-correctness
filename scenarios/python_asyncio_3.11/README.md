## Basic `asyncio` check

### What it does

This checks the correctness of a basic `asyncio`-based script.

Default execution time is 3 seconds.

The script has

- `my_coroutine`: idle coroutine that just calls `asyncio.sleep`
- `long_computation`: CPU-heavy coroutine that computes something
- `async_def`: entry point, which
  - Creates a `my_coroutine` Task (running in the background) for `EXEC_TIME / 2`, name is `short_task`
  - Awaits `long_computation` for `EXEC_TIME / 3`
  - Gathers the `short_task` (partially completed) and a new Task (`Task-3`) for `my_coroutine` for `EXEC_TIME`

### Expected

The expected Profile should have the following stacks:

- `my_coroutine;sleep`: ~2 seconds in `short_task`
- `main;long_computation`: ~1 second in `Task-1`
- `main;my_coroutine`: ~1 second in `short_task`
- `main;my_coroutine`: ~3 seconds in `Task-3`

### TODO

- [ ] Investigate why wall-time shows ~1.5 seconds instead of expected 2 seconds
