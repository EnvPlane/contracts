# Historical local tag refs differ from canonical remote

Status: open local checkout hygiene issue; no release tags changed.

## Evidence

A read-only `git fetch origin main --tags` on 2026-10-10 rejected local tag
collisions for v0.1.18, v0.1.19, v0.1.42, v0.1.44, v0.1.45, v0.1.73 and v0.1.76.
For example, local v0.1.18 and v0.1.19 both peel to
`c967e6f218f3a8fe8a78377de2c44b3cbfce6b9a`; canonical remote v0.1.18 peels to
`d55cd530aa717ef5e7f1e00c00d84db7c18e808d` and v0.1.19 to
`255f336cbb91fab33ce8899b552d077988759dd5`. The fetch did correctly obtain
canonical v0.1.110 and v0.1.111; latest v0.1.111 remains bound to e69da406.

This does not prove a canonical remote tag was moved. A fresh hosted checkout
does not inherit these pre-existing local refs. Publisher fetch must fail closed
on mismatched immutable tags; it must never force replacement to hide this.

## Implementation prompt

Audit historical local tag provenance against canonical remote tag objects and
peeled source SHAs. Preserve recovery evidence. Prefer a fresh separate checkout
for publication diagnostics. Do not delete, move, force-fetch or republish any
tag in the shared checkout or remote without explicit targeted authorization.
Keep module/domain/OpenAPI and all consumer source untouched. Record the
resolution separately from the publisher fix; this task does not authorize
historical release repair.
