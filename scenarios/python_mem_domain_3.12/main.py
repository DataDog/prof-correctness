from ddtrace.profiling import Profiler


def allocate_mem_domain_buffers() -> list[object]:
    # On 3.12+, list multiplication allocates the ob_item pointer array in
    # PYMEM_DOMAIN_MEM — the allocation class mem_domain tracking targets.
    return [[None] * 4096 for _ in range(256)]


if __name__ == "__main__":
    prof = Profiler()
    prof.start()
    _objs = allocate_mem_domain_buffers()
    prof.stop()
