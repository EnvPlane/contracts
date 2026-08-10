#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
policy="$repo_root/docs/adr/0001-open-core-and-enterprise-boundary.md"

for repository in contracts agent runner webhook bootstrap gitops control-plane frontend deploy; do
  rg -q "\`$repository\`" "$policy"
done

for section in "Private Enterprise modules" "must not import private modules" "Legal and release checklist" "read, delete, cleanup or export"; do
  rg -q "$section" "$policy"
done
