#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
inventory="$repo_root/docs/legacy-machine-identifiers.json"

usage() {
  printf 'usage: %s [--classify IDENTIFIER|--category CATEGORY|--scan-root DIRECTORY]\n' "$0" >&2
}

test -f "$inventory"

jq -e '
  .schemaVersion == 2 and
  .canonicalProductName == "EnvPlane" and
  (.earliestRemovalMajor | (type == "number" and . >= 2)) and
  (.deprecationPolicy | type == "string" and length > 0) and
  (.categories | type == "array" and length == 8) and
  ([.categories[].id] | sort) == [
    "cli_config_paths_and_commands",
    "environment_variables",
    "external_urls",
    "go_module_paths",
    "kubernetes_names_labels_and_api_groups",
    "oci_helm_release_and_runtime_names",
    "persistence_metrics_and_queue_names",
    "webhook_commands"
  ] and
  ([.categories[].id] | unique | length == 8) and
  all(.categories[];
    (.id | type == "string" and length > 0) and
    (.description | type == "string" and length > 0) and
    (.match | type == "string" and length > 0) and
    (.ownerRepos | type == "array" and length > 0) and
    (.compatibilityRisk | type == "string" and length > 0) and
    (.readPath | type == "string" and length > 0) and
    (.writePath | type == "string" and length > 0) and
    (.migrationStrategy | type == "string" and length > 0) and
    (.earliestRemovalMajor | (type == "number" and . >= 2)) and
    (.identifiers | type == "array" and length > 0 and all(type == "string" and length > 0))
  )
' "$inventory" >/dev/null

classify() {
  local identifier="$1"
  if jq -e --arg identifier "$identifier" \
    'any(.categories[]; .match as $pattern | ($identifier | test($pattern)))' \
    "$inventory" >/dev/null; then
    return 0
  fi
  printf 'unclassified legacy machine identifier: %s\n' "$identifier" >&2
  return 1
}

scan_root() {
  local root="$1"
  test -d "$root"
  local identifiers
  identifiers="$(rg --no-heading --no-filename --only-matching --replace '$0' \
    --glob '!.git/**' --glob '!node_modules/**' --glob '!dist/**' --glob '!build/**' --glob '!vendor/**' \
    -e 'ENVPLANE_[A-Z0-9_]+' \
    -e 'github\.com/envplane/[A-Za-z0-9._/-]+' \
    -e 'ghcr\.io/envplane/[A-Za-z0-9._/-]+' \
    -e 'envplane\.io/[A-Za-z0-9._/-]+' \
    -e '/envplane[A-Za-z0-9._/-]*' \
    -e 'envplane[-_][A-Za-z0-9._-]+' \
    -e 'envplane~[A-Za-z0-9._-]+' \
    -e 'https?://[A-Za-z0-9._-]*envplane[A-Za-z0-9._:/?&=#%+~-]*' \
    "$root" | sort -u || true)"
  while IFS= read -r identifier; do
    test -n "$identifier" || continue
    classify "$identifier"
  done <<< "$identifiers"
}

case "${1:-}" in
  "") ;;
  --classify)
    test "${2:-}" != ""
    classify "$2"
    ;;
  --category)
    test "${2:-}" != ""
    jq -e --arg category "$2" 'any(.categories[]; .id == $category)' "$inventory" >/dev/null || {
      printf 'unaccounted legacy identifier category: %s\n' "$2" >&2
      exit 1
    }
    ;;
  --scan-root)
    test "${2:-}" != ""
    scan_root "$2"
    ;;
  *)
    usage
    exit 2
    ;;
esac
