#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
generated="$repo_root/sdk/go/envplanesdk/openapi_contract.gen.go"
temporary="$(mktemp)"
trap 'rm -f "$temporary"' EXIT

cd "$repo_root"
go run ./cmd/sdkgen -spec openapi/openapi.json -out "$temporary"
if ! cmp -s "$temporary" "$generated"; then
  echo "generated SDK contract metadata is stale; run go generate ./sdk/go/envplanesdk" >&2
  diff -u "$generated" "$temporary" || true
  exit 1
fi
