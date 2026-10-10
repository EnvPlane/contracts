#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
expected="${ENVPLANE_CONTRACTS_VERSION:-}"
status=0
[[ -d "$root" ]] || { echo "workspace root does not exist: $root" >&2; exit 2; }
while IFS= read -r -d '' file; do
  version="$(awk '$1 == "github.com/envplane/contracts" && $2 != "=>" { print $2; exit } $1 == "require" && $2 == "github.com/envplane/contracts" { print $3; exit }' "$file")"
  [[ -n "$version" ]] || continue
  if [[ -z "$expected" ]]; then
    echo "ENVPLANE_CONTRACTS_VERSION must specify the verified published tag when checking consumers" >&2
    exit 2
  fi
  if [[ "$version" != "$expected" ]]; then
    echo "$file pins $version, expected $expected" >&2
    status=1
  fi
done < <(find "$root" -type d \( -name .git -o -name node_modules -o -name vendor \) -prune -o -type f -name go.mod -print0)
exit "$status"
