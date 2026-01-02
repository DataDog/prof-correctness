import os
from time import time

x = 0
i = 0


def main():
    global x, i
    EXECUTION_TIME_SEC = int(
        os.getenv("EXECUTION_TIME_SEC", "10")
    )  # defaults to 10 if not set
    end = time() + EXECUTION_TIME_SEC
    while time() < end:
        a()
        b()
    # We add a print to prevent optimization that could turn this into a no-op program
    print(x)


def a():
    global x, i
    i = 0
    while i < 1000000:
        x += i
        i += 1


def b():
    global x, i
    i = 0
    while i < 2000000:
        x += i
        i += 1


main()
