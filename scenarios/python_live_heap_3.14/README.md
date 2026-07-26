## Live-Heap Profiling (3.14 baseline)

Validates that the Datadog Python profiler's **persistent live-heap profile**
correctly reports the set of live (still-allocated) sampled objects. This is the
**3.14 baseline** half of the `3.14 -> 3.15` migration pair (see
`python_live_heap_3.15`).

Unlike `alloc-space`/`alloc-samples` (which count every allocation over the
interval), the live-heap profile is a running snapshot of what is *currently
live*: allocations are added when sampled and subtracted when freed, and the
snapshot is exported non-destructively on every upload. It is attached as a
second pprof (`<prefix>.<pid>.<seq>.heap.pprof`) alongside the primary profile.

## Test Application

`main.py` retains a known live set for the whole run, split into two
distinctly-named call sites that allocate **equal-size** objects (16 KiB) and
differ only in count:

- `retain_major` - 1,600 objects (~25 MiB, ~80% of the live set)
- `retain_minor` -   400 objects (~6 MiB, ~20% of the live set)

Equal sizes mean the 80/20 split holds for both metrics the live-heap profile
reports: `heap-space` (live bytes = count x size) and `heap-live-samples` (live
object count). It uses `bytes` (`PyObject_Malloc`, the OBJ allocator domain) so
the heap profiler tracks the objects identically across versions, independent of
the `DD_PROFILING_MEMORY_MEM_DOMAIN_ENABLED` toggle (`bytearray` moved OBJ -> MEM
in 3.13). The objects are held alive while the process idles through several
upload intervals (`DD_PROFILING_UPLOAD_INTERVAL=5`), so each exported heap
snapshot contains the full live set.

## Expected Profile

Assertions run against the heap snapshot pprof only (selected with
`pprof-regex`). `DD_PROFILING_HEAP_SAMPLE_SIZE=16384` (== object size) yields
roughly one sample per object, keeping the per-stack proportions stable.

- `heap-space` (live bytes):
  - `^<module>;Target.run;Target.retain_major$` ~= 80%
  - `^<module>;Target.run;Target.retain_minor$` ~= 20%
- `heap-live-samples` (live object count): same ~80/20 split.

`allow_first_profile_failure` tolerates the first snapshot, which may be taken
before the live set is fully built.

## Notes

The persistent live-heap profile is recent, so this scenario runs against a
ddtrace build that includes it (via `DDTRACE_INSTALL_URL` pointing at a
dd-trace-py S3 wheel) rather than a PyPI release. It is excluded from
prof-correctness `main` CI until the feature ships in a release.
