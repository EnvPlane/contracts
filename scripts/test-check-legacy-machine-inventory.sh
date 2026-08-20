#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
check="$repo_root/scripts/check-legacy-machine-inventory.sh"

"$check"
"$check" --classify ENVPLANE_API_TOKEN
"$check" --classify github.com/envplane/contracts
"$check" --classify envplane.io/environment-id
"$check" --classify /envplane

if "$check" --classify unrelated_legacy_identifier; then
  echo "inventory accepted an unclassified identifier" >&2
  exit 1
fi
