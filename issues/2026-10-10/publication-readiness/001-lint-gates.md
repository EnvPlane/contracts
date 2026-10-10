# PVC/MySQL staged contracts lint gates

Status: fixed locally; verify on frozen remote candidate after main acceptance.

Findings: pinned golangci-lint v2.12.2 flagged QF1001 on the PVC access-mode
predicate and G101 on a MySQL test fixture's Kubernetes Secret key name.
The later independent-root-key change also triggers G101 on the fixed root
Secret key-name constant; its suppression is scoped to that metadata declaration.

Implementation prompt: preserve predicate semantics while applying De Morgan's
law. Scope the G101 suppression to the fixture line and explain that PasswordKey
is a key name, not a credential. Never suppress production credential checks or
weaken CI. Run tests/race/vet/lint and require the frozen candidate's CI green.
