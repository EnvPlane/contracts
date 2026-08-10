#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
inventory="$repo_root/docs/legacy-machine-identifiers.json"

if [[ ! -f "$inventory" ]]; then
  echo "missing legacy machine identifier inventory" >&2
  exit 2
fi

jq -e '
  .schemaVersion == 1 and
  .canonicalProductName == "EnvPlane" and
  (.deprecationPolicy | length > 0) and
  (.categories | type == "array" and length >= 6) and
  (([.categories[] | .id] | length) == ([.categories[] | .id] | unique | length)) and
  all(.categories[]; ((.id | type) == "string" and (.id | length) > 0) and ((.match | type) == "string" and (.match | length) > 0) and ((.owners | type) == "array" and (.owners | length) > 0) and ((.risk | type) == "string" and (.risk | length) > 0) and ((.migration | type) == "string" and (.migration | length) > 0))
' "$inventory" >/dev/null

if [[ "${1:-}" == "--classify" ]]; then
  identifier="${2:-}"
  if [[ -z "$identifier" ]]; then
    echo "usage: $0 --classify <identifier>" >&2
    exit 2
  fi
  if ! jq -e --arg identifier "$identifier" 'any(.categories[]; .match as $pattern | ($identifier | test($pattern)))' "$inventory" >/dev/null; then
    echo "unclassified legacy machine identifier: $identifier" >&2
    exit 1
  fi
fi
