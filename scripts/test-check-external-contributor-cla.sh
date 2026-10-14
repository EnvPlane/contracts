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

"$check" --association CONTRIBUTOR --labels "dependencies,go" \
  --author-login 'dependabot[bot]' --author-id 49699333 --author-type Bot

expect_rejection() {
  if "$check" --association CONTRIBUTOR --labels "dependencies,go" "$@"; then
    echo "unverified dependency contributor passed without a CLA" >&2
    exit 1
  fi
}
expect_rejection
expect_rejection --author-login 'dependabot[bot]' --author-id 1 --author-type Bot
expect_rejection --author-login 'dependabot[bot]' --author-id 49699333 --author-type User
expect_rejection --author-login 'other[bot]' --author-id 49699333 --author-type Bot
expect_rejection --author-login 'dependabot[bot]' --author-type Bot

PR_AUTHOR_LOGIN='dependabot[bot]' PR_AUTHOR_ID=49699333 PR_AUTHOR_TYPE=Bot \
  "$check" --association CONTRIBUTOR --labels "dependencies,go"
