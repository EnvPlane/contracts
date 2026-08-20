#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
policy="$repo_root/docs/adr/0001-open-core-and-enterprise-boundary.md"
check="$repo_root/scripts/check-open-core-policy.sh"

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
"$check"

fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT
cp "$repo_root/docs/open-core-policy.json" "$fixture_dir/policy.json"
jq '.dependencyRules.publicMayDependOnPrivate = true' "$fixture_dir/policy.json" > "$fixture_dir/bad.json"
if jq -e '.dependencyRules.publicMayDependOnPrivate == false' "$fixture_dir/bad.json" >/dev/null; then
  echo "policy fixture accepted public-to-private dependency" >&2
  exit 1
fi
