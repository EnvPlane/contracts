#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(cd "$(dirname "$0")/../.." && pwd)}"
repos=(contracts control-plane runner agent gitops bootstrap webhook frontend deploy)
list="$(cd "$(dirname "$0")" && pwd)/shared-files.txt"
status=0

while IFS= read -r relative; do
  [[ -z "$relative" || "$relative" == \#* ]] && continue
  matches=()
  for repo in "${repos[@]}"; do
    path="$root/$repo/$relative"
    [[ -f "$path" ]] && matches+=("$repo")
  done
  if ((${#matches[@]} > 1)); then
    echo "duplicate canonical file $relative: ${matches[*]}" >&2
    status=1
  fi
done < "$list"

exit "$status"
