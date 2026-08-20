#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
expected="${ENVPLANE_CONTRACTS_VERSION:-v0.1.3}"
status=0
for module in control-plane runner agent gitops bootstrap webhook; do
  file="$root/$module/go.mod"
  [[ -f "$file" ]] || continue
  version="$(awk '$1 == "github.com/envplane/contracts" && $2 != "=>" { print $2; exit } $1 == "require" && $2 == "github.com/envplane/contracts" { print $3; exit }' "$file")"
  if [[ "$version" != "$expected" ]]; then
    echo "$module pins $version, expected $expected" >&2
    status=1
  fi
done
exit "$status"
