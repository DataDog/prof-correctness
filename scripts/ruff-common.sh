# Shared setup for scripts/lint and scripts/format.
ruff_common_setup() {
  RUFF_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  cd "$RUFF_ROOT"

  if ! command -v ruff >/dev/null 2>&1; then
    echo "ruff not found; install with: pip install -r $RUFF_ROOT/requirements-dev.txt" >&2
    exit 1
  fi
}

ruff_resolve_targets() {
  RUFF_TARGETS=("$@")
  if [ "${#RUFF_TARGETS[@]}" -eq 0 ]; then
    RUFF_TARGETS=(".")
  fi
}
