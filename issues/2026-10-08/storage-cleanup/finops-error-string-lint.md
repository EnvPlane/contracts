# FinOps endpoint validation lint regression

Status: fixed locally; no push.

The full contracts lint run found two ST1005 violations in existing remote
FinOps endpoint error strings. Validation semantics and error conditions are
unchanged; messages now start with lowercase `metrics endpoint`.

Codex prompt: preserve credential-free HTTPS and exact origin validation;
keep returned error strings lowercase and run domain tests plus golangci-lint.
