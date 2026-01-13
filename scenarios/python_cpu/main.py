import os
from time import time


class CPUBurner:
    def __init__(self) -> None:
        self.x = 0
        self.i = 0

    def a(self) -> None:
        self.i = 0
        while self.i < 1000000:
            self.x += self.i
            self.i += 1

    def b(self) -> None:
        self.i = 0
        while self.i < 2000000:
            self.x += self.i
            self.i += 1

    def main(self) -> None:
        execution_time_sec = float(os.getenv("EXECUTION_TIME_SEC", "10"))
        end = time() + execution_time_sec
        while time() < end:
            self.a()
            self.b()


CPUBurner().main()
