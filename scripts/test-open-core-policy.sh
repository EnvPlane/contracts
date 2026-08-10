#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
policy="$repo_root/docs/adr/0001-open-core-and-enterprise-boundary.md"

contains() {
  if command -v rg >/dev/null 2>&1; then
    rg -q -- "$1" "$policy"
  else
    grep -Fq -- "$1" "$policy"
  fi
}

for repository in contracts agent runner webhook bootstrap gitops control-plane frontend deploy; do
  contains "\`$repository\`"
done

for section in "Private Enterprise modules" "must not import private modules" "Legal and release checklist" "read, delete, cleanup or export"; do
  contains "$section"
done
"$(dirname "$0")/check-module-boundary.sh" "$(git rev-parse --show-toplevel)"
