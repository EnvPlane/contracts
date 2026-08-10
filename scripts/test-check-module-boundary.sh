#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
check="$repo_root/scripts/check-module-boundary.sh"
fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT

mkdir -p "$fixture_dir/public"
printf '%s\n' 'package public' 'import "github.com/envpilot/contracts/domain"' > "$fixture_dir/public/allowed.go"
"$check" "$fixture_dir"

printf '%s\n' 'package public' 'import "example.com/vendor/enterprise/license"' > "$fixture_dir/public/forbidden.go"
if "$check" "$fixture_dir"; then
  echo "module boundary accepted a private Go import" >&2
  exit 1
fi
rm "$fixture_dir/public/forbidden.go"

mkdir "$fixture_dir/enterprise"
if "$check" "$fixture_dir"; then
  echo "module boundary accepted an enterprise source directory" >&2
  exit 1
fi
