#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
check="$repo_root/scripts/check-legacy-machine-inventory.sh"

"$check"
for category in \
  environment_variables \
  go_module_paths \
  oci_helm_release_and_runtime_names \
  kubernetes_names_labels_and_api_groups \
  cli_config_paths_and_commands \
  webhook_commands \
  persistence_metrics_and_queue_names \
  external_urls; do
  "$check" --category "$category"
done

"$check" --classify ENVPLANE_API_TOKEN
"$check" --classify github.com/envplane/contracts
"$check" --classify envplane.io/environment-id
"$check" --classify /envplane/payment
"$check" --classify https://envplane.example.test/health

if "$check" --category future_untracked_category; then
  echo "inventory accepted an unaccounted category" >&2
  exit 1
fi
untracked="envplane""~untracked"
if "$check" --classify "$untracked"; then
  echo "inventory accepted an unclassified identifier" >&2
  exit 1
fi

fixture="$(mktemp -d)"
trap 'rm -rf "$fixture"' EXIT
cat > "$fixture/known.txt" <<'EOF'
ENVPLANE_API_TOKEN
github.com/envplane/contracts
envplane.io/environment-id
https://envplane.example.test
EOF
"$check" --scan-root "$fixture"
printf '\n%s\n' "$untracked" >> "$fixture/known.txt"
if "$check" --scan-root "$fixture"; then
  echo "inventory scan accepted an unclassified identifier" >&2
  exit 1
fi
