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
        execution_time_sec = int(os.getenv("EXECUTION_TIME_SEC", "10")) # defaults to 10 if not set
        end = time() + execution_time_sec
        while time() < end:
            self.a()
            self.b()

        # We add a print to prevent optimization that could turn this into a no-op program
        print(self.x)


CPUBurner().main()
