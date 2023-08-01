import os
from time import time

x = 0
i = 0

def main():
    global x, i
    execution_time = int(os.getenv("EXECUTION_TIME", "10")) # defaults to 10 if not set
    end = time() + execution_time
    while time() < end:
        a()
        b()
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
