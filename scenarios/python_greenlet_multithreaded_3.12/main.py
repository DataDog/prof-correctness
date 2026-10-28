"""Test scenario for greenlet profiling with multiple threads."""

import random
import threading
import time

import greenlet


class MathWorker:
    def __init__(self) -> None:
        self.end_time = time.monotonic() + 10

        self.fib_greenlet = greenlet.greenlet(self.work_fibonacci)
        self.prime_greenlet = greenlet.greenlet(self.work_prime)

        self.last_fib = 0
        self.last_prime = 0

    def do_work(self) -> None:
        self.fib_greenlet.switch()
        print(f"Fibonacci: {self.last_fib}, Prime: {self.last_prime}")

    def work_fibonacci(self) -> None:
        while True:
            # Find the next Fibonacci number
            self.last_fib = self.last_fib + 1

            a, b = 0, 1
            for _ in range(self.last_fib):
                a, b = b, a + b

            if self.end_time < time.monotonic():
                return

            self.prime_greenlet.switch()

    def work_prime(self) -> None:
        while True:
            for _ in range(5):
                # Find the next prime number after self.last_prime
                candidate = self.last_prime + 1
                while True:
                    is_prime = candidate > 1 and all(candidate % i != 0 for i in range(2, int(candidate**0.5) + 1))
                    if is_prime:
                        self.last_prime = candidate
                        break
                    candidate += 1

            if self.end_time < time.monotonic():
                return

            self.fib_greenlet.switch()


class UtilWorker:
    def __init__(self) -> None:
        self.end_time = time.monotonic() + 10

        self.string_greenlet = greenlet.greenlet(self.work_string)
        self.list_greenlet = greenlet.greenlet(self.work_list)

        self.string_iterations = 0
        self.list_iterations = 0

    def do_work(self) -> None:
        self.string_greenlet.switch()
        print(f"String iterations: {self.string_iterations}, List iterations: {self.list_iterations}")

    def work_string(self) -> None:
        while True:
            result = ""
            for _ in range(100):
                result += str(random.randint(0, 100))  # noqa: S311

            self.string_iterations += 1

            if self.end_time < time.monotonic():
                return

            self.list_greenlet.switch()

    def work_list(self) -> None:
        while True:
            sums = [1]
            for _ in range(100):
                sums.append(sum(sums))

            self.list_iterations += 1

            if self.end_time < time.monotonic():
                return

            self.string_greenlet.switch()


def main() -> None:
    t = threading.Thread(target=lambda: MathWorker().do_work(), name="MathWorker")
    u = threading.Thread(target=lambda: UtilWorker().do_work(), name="UtilWorker")

    t.start()
    u.start()

    t.join()
    u.join()


if __name__ == "__main__":
    main()
