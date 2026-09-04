#!/bin/sh
set -eu

# Pre-install ddtrace from a dd-trace-py wheel when DDTRACE_INSTALL_URL is set.
if [ -n "${DDTRACE_INSTALL_URL:-}" ]; then
  apt-get update
  apt-get install -y --no-install-recommends curl ca-certificates
  rm -rf /var/lib/apt/lists/*
  curl -fsSLo /tmp/ddtrace-install.sh "$DDTRACE_INSTALL_URL"
  # Pinned cp315 wheels still declare Requires-Python <3.15 until #19942 is in that SHA.
  PIP_IGNORE_REQUIRES_PYTHON=1 bash /tmp/ddtrace-install.sh
fi
