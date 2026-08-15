#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if rg -n 'db:"' "$root/domain" --glob '*.go'; then
  echo "contracts/domain must not contain database tags" >&2
  exit 1
fi
echo "contracts/domain boundary is clean"
