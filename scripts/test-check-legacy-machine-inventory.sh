#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
check="$repo_root/scripts/check-legacy-machine-inventory.sh"

"$check"
"$check" --classify ENVPILOT_API_TOKEN
"$check" --classify github.com/envpilot/contracts
"$check" --classify envpilot.io/environment-id
"$check" --classify /envpilot

if "$check" --classify unrelated_legacy_identifier; then
  echo "inventory accepted an unclassified identifier" >&2
  exit 1
fi
