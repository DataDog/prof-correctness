import os
from time import time

import torch


def light_compute(a: torch.Tensor, b: torch.Tensor) -> None:
    for _ in range(20):
        torch.mm(a, b)


def heavy_compute(a: torch.Tensor, b: torch.Tensor) -> None:
    for _ in range(40):
        torch.mm(a, b)


def main() -> None:
    torch.set_num_threads(1)
    a = torch.randn(300, 300)
    b = torch.randn(300, 300)
    execution_time_sec = float(os.getenv("EXECUTION_TIME_SEC", "10"))
    end = time() + execution_time_sec
    while time() < end:
        light_compute(a, b)
        heavy_compute(a, b)


main()
