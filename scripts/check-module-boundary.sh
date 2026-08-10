#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
# Public modules may only expose public contracts. Keep the guard deliberately
# conservative: private/enterprise paths must never appear in Go imports.
if rg -n --glob '*.go' '"[^" ]*/(private|enterprise)(/|")' "$root"; then
  echo "private module import detected in public module" >&2
  exit 1
fi
echo "module boundary check passed"
