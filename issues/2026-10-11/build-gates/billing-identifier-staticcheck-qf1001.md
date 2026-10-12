# Billing identifier allowlist fails staticcheck QF1001

Status: fixed; local verification passed.

The contracts lint gate rejects the negated compound ASCII allowlist in
`domain/billing_webhook_verification.go` with QF1001. This is a code-style gate
failure, not a Go toolchain download failure.

## Implementation prompt

Rewrite the identifier validation using a positive allowlist without weakening
the ASCII-only, nonempty or maximum-length constraints. Add regression tests
for allowed characters, range boundaries, Unicode, separators and length.
Run the full contracts tests and `scripts/golangci-lint.sh`. Do not suppress
the lint rule or change the public contract.

## Fix

Use a switch with allowed ASCII cases and a rejecting default. Add direct
identifier regression coverage; no schema or signature behavior changes.

## Verification

- `GOWORK=off GOTOOLCHAIN=go1.26.9 go test -race ./...`: passed.
- `GOWORK=off GOTOOLCHAIN=go1.26.9 scripts/golangci-lint.sh`: 0 issues.
- Brand diff gate and `git diff --check`: passed.
- Hosted CI will be confirmed separately after push.
