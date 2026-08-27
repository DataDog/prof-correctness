# Contributing

Community contributions to `prof-correctness` are welcome. See below for some basic guidelines.

## Highlighting setups in which profiling should work better

If you want to highlight a use case on which you would like profiling to work better, this is a good place to make a PR.
Checkout the readme for some guidelines on how to create new tests.

## Choosing expected values

The JSON next to a scenario states what the workload is **supposed** to produce, not what the last CI run happened to observe.

1. Prefer an absolute or rate `value` derived from the workload constants (time budgets, byte ratios, thread counts, durations). Reserve `percent` for cases where no absolute quantity exists.
2. Run the scenario (locally or in CI) and use the captured JSON under `./data/` to confirm the derivation and to size `error_margin` from measured spread.
3. If no closed form exists, put the reason in the profile `note` and harvest a distribution (`TestFlakiness` / `FLAKINESS_RUNS`; inspect the captured JSON dumps under `./data/`) rather than locking a single sample.

See README "Describe the expected output" for the full convention.
