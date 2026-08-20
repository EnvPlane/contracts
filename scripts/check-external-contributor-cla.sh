#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'usage: %s --association ASSOCIATION --labels COMMA_SEPARATED_LABELS\n' "$0" >&2
}

association=""
labels=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --association)
      association="${2:-}"
      shift 2
      ;;
    --labels)
      labels="${2:-}"
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "$association" ]]; then
  usage
  exit 2
fi

case "$association" in
  OWNER|MEMBER|COLLABORATOR)
    echo "contributor agreement check passed for trusted repository association"
    exit 0
    ;;
esac

IFS=',' read -r -a label_list <<< "$labels"
for label in "${label_list[@]}"; do
  if [[ "$label" == "cla-signed" ]]; then
    echo "contributor agreement check passed"
    exit 0
  fi
done

echo "external pull request requires a verified individual CLA and the cla-signed label" >&2
exit 1
