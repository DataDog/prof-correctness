from ddtrace.profiling import Profiler


class Target:
    def __init__(self) -> None:
        self.memory: list[bytearray | None] = []
        self.index: int = 0
        self.grow_list(target=int(1e6))

    def run(self) -> None:
        while self.memory[-1] is None:
            self.allocate_memory_1(1024)
            self.allocate_memory_2(1024)

    def grow_list(self, target: int) -> None:
        self.memory = [None for _ in range(target)]

    def allocate_memory_1(self, size: int) -> None:
        self.memory[self.index] = bytearray(size)
        self.index += 1

    def allocate_memory_2(self, size: int) -> None:
        self.memory[self.index] = bytearray(3 * size)
        self.index += 1


def main() -> None:
    prof: Profiler = Profiler()
    prof.start()

    target: Target = Target()
    target.run()

    prof.stop()


if __name__ == "__main__":
    main()
