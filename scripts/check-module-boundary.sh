#!/usr/bin/env bash
set -euo pipefail

root="${1:-.}"
if [[ ! -d "$root" ]]; then
  echo "module boundary root is not a directory: $root" >&2
  exit 2
fi

private_dir="$(find "$root" -type d \( -name private -o -name enterprise \) -not -path '*/.git/*' -print -quit)"
if [[ -n "$private_dir" ]]; then
  echo "private module directory detected in public module: $private_dir" >&2
  exit 1
fi

if find "$root" -type l -not -path '*/.git/*' -print -quit | grep -q .; then
  echo "symbolic links are not allowed in the public contract module" >&2
  exit 1
fi

failed=0
if command -v rg >/dev/null 2>&1; then
  if rg -n --glob '*.go' '"[^" ]*/(private|enterprise)(/|")' "$root"; then failed=1; fi
  if rg -n --glob 'go.mod' '^[[:space:]]*(require|replace)?[[:space:]]*[^[:space:]]*/(private|enterprise)([[:space:]/]|$)' "$root"; then failed=1; fi
  if rg -n --glob '*.{js,jsx,ts,tsx,mjs,cjs}' '(from[[:space:]]+|require\(|import\()["'\''][^"'\'']*/(private|enterprise)(/|["'\''])' "$root"; then failed=1; fi
else
  if find "$root" -type f -name '*.go' -not -path '*/.git/*' -exec grep -HnE '"[^" ]*/(private|enterprise)(/|")' {} +; then failed=1; fi
  if find "$root" -type f -name 'go.mod' -not -path '*/.git/*' -exec grep -HnE '^[[:space:]]*(require|replace)?[[:space:]]*[^[:space:]]*/(private|enterprise)([[:space:]/]|$)' {} +; then failed=1; fi
  if find "$root" -type f \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.mjs' -o -name '*.cjs' \) -not -path '*/.git/*' -exec grep -HnE '(from[[:space:]]+|require\(|import\()["'\''][^"'\'']*/(private|enterprise)(/|["'\''])' {} +; then failed=1; fi
fi

if [[ "$failed" -ne 0 ]]; then
  echo "private module dependency detected in public module" >&2
  exit 1
fi
echo "module boundary check passed"
