from ddtrace.profiling import Profiler


def allocate_mem_domain_buffers() -> list[object]:
    # On Python 3.13+, bytearray's internal buffer is allocated in PYMEM_DOMAIN_MEM.
    return [bytearray(4096) for _ in range(256)]


if __name__ == "__main__":
    prof = Profiler()
    prof.start()
    _objs = allocate_mem_domain_buffers()
    prof.stop()
