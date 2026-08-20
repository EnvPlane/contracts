#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
check="$repo_root/scripts/check-external-contributor-cla.sh"

"$check" --association OWNER --labels ""
"$check" --association NONE --labels "documentation,cla-signed"

if "$check" --association FIRST_TIME_CONTRIBUTOR --labels "documentation"; then
  echo "external contributor passed without a verified CLA label" >&2
  exit 1
fi

if "$check" --association UNKNOWN --labels ""; then
  echo "unknown contributor association passed without a verified CLA label" >&2
  exit 1
fi
