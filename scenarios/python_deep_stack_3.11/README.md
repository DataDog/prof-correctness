# python_deep_stack_3.11

Verifies that the stack unwinder correctly captures deep call stacks. Every
other scenario has shallow stacks (fewer than ~10 frames), so the unwinder's
depth loop, its cycle-detection guard, and the frame cache under hundreds of
repeated identical frames are otherwise untested.

The workload recurses 400 levels deep through `recurse` and then burns CPU in
`burn` at the leaf for the full duration.

## Expected behavior

- **cpu-time**: nearly all CPU is attributed to a stack of the form
  `<N frames omitted>;recurse;...;recurse;burn`. The exporter keeps only the
  innermost ~64 frames and collapses the rest into a `<N frames omitted>`
  marker, so the marker's presence (plus the dozens of retained consecutive
  `recurse` frames ending in `burn`) proves the sampler walked the full deep
  stack and counted the omitted frames.
