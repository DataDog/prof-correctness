# python_pytorch_profiler_3.11

Correctness check for the dd-trace-py **PyTorch profiler integration**
(`ddtrace/profiling/collector/pytorch.py`).

## What it checks

When `DD_PROFILING_PYTORCH_ENABLED` is set, dd-trace-py monkey-patches
`torch.profiler.profile` and converts each recorded `FunctionEvent` into a
Datadog profiler sample. The integration reconstructs the operator call tree by
walking each event's `cpu_parent` chain, so a nested operator must appear as a
nested stack rooted at the `PYTORCH_DeviceType.CPU` pseudo-frame:

```
PYTORCH_DeviceType.CPU;aten::linear;aten::addmm
```

Before the call-tree reconstruction, every operator was emitted as a flat
two-frame stack (`PYTORCH_DeviceType.CPU;aten::addmm`), which produced a flat
flame graph. This check asserts the nested form, so it fails against the old
flat behavior and passes once the operator nesting is emitted.

## How it works

- `main.py` repeatedly runs a `torch.nn.Linear` forward inside a
  `torch.profiler.profile(activities=[CPU])` context. `aten::linear` is the
  parent operator and `aten::addmm` (the matrix multiply) is its dominant child.
- The Python stack sampler is disabled (`DD_PROFILING_STACK_ENABLED=false`) so
  the `cpu-time` profile contains only the torch-operator samples emitted by the
  integration. This keeps the per-operator percentages stable.

## Running

```sh
TEST_SCENARIOS="python_pytorch_profiler.*" go test -v -run TestScenarios
```

To validate against a locally-built dd-trace-py wheel, pass its install script:

```sh
DDTRACE_INSTALL_URL="https://.../install.sh" \
  TEST_SCENARIOS="python_pytorch_profiler.*" go test -v -run TestScenarios
```

> The `error_margin` is intentionally wide: the structural regex (a parent
> `aten::linear` containing a child matmul op) is the correctness signal, while
> the percentage only guards against the operator carrying no cpu-time (e.g. the
> old flat behavior, where the nested stack is absent and matches 0%).
