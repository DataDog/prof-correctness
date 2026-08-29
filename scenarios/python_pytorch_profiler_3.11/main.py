import os
from time import time

import torch
from torch import nn
from torch.profiler import ProfilerActivity


def main() -> None:
    # Single-threaded BLAS keeps the per-operator timings stable across runs.
    torch.set_num_threads(1)

    # A Linear forward records a parent ``aten::linear`` operator whose dominant
    # child is ``aten::addmm`` (the matrix multiply). This parent/child nesting
    # is exactly what the dd-trace-py torch profiler integration reconstructs by
    # walking each event's ``cpu_parent`` chain.
    layer = nn.Linear(1024, 1024)
    x = torch.randn(512, 1024)

    execution_time_sec = float(os.getenv("EXECUTION_TIME_SEC", "10"))
    end = time() + execution_time_sec
    with torch.no_grad():
        while time() < end:
            # Re-enter the profiler in batches so ``on_trace_ready`` fires
            # periodically (the integration emits samples on each trace) and the
            # number of buffered events stays bounded.
            with torch.profiler.profile(activities=[ProfilerActivity.CPU]):
                for _ in range(50):
                    layer(x)


main()
