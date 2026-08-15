#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(cd "$(dirname "$0")/../.." && pwd)}"
repos=(contracts control-plane runner agent gitops bootstrap webhook frontend deploy)
list="$(cd "$(dirname "$0")" && pwd)/shared-files.txt"
status=0

while IFS= read -r relative; do
  [[ -z "$relative" || "$relative" == \#* ]] && continue
  baseline=""
  baseline_repo=""
  for repo in "${repos[@]}"; do
    path="$root/$repo/$relative"
    [[ -f "$path" ]] || continue
    digest="$(shasum -a 256 "$path" | awk '{print $1}')"
    if [[ -z "$baseline" ]]; then
      baseline="$digest"
      baseline_repo="$repo"
    elif [[ "$digest" != "$baseline" ]]; then
      echo "shared file drift in $relative: $baseline_repo != $repo" >&2
      status=1
    fi
  done
done < "$list"

exit "$status"
