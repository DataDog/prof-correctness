#!/usr/bin/env bash

ddprof "$@"
exit_code=$?

# retrieve the pid of profiler and wait for it to finish
pid=$(pgrep ddprof | sort -n | head -n1)
[ -n "$pid" ] && tail -f /dev/null --pid "$pid"

exit "$exit_code"
