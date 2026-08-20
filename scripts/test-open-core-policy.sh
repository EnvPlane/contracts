#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
check="$repo_root/scripts/check-open-core-policy.sh"
fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT

"$(dirname "$0")/check-module-boundary.sh" "$repo_root"
"$check"

cp "$repo_root/docs/open-core-policy.json" "$fixture_dir/policy.json"
jq '.dependencyRules.publicMayDependOnPrivate = true' "$fixture_dir/policy.json" > "$fixture_dir/bad-dependency.json"
if jq -e '.dependencyRules.publicMayDependOnPrivate == false' "$fixture_dir/bad-dependency.json" >/dev/null; then
  echo "policy fixture accepted public-to-commercial dependency" >&2
  exit 1
fi

jq '.commercialRepositories[0].license = "Apache-2.0"' "$fixture_dir/policy.json" > "$fixture_dir/bad-license.json"
if jq -e 'any(.commercialRepositories[]; .name == "control-plane" and .license == "BSL-1.1")' "$fixture_dir/bad-license.json" >/dev/null; then
  echo "policy fixture accepted incorrect control-plane license" >&2
  exit 1
fi
