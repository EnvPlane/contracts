# PVC copy: consumer execution and lifecycle verification

Status: integration handoff; outside this contracts-only change.

## Required implementation prompt for Codex

Integrate `domain.PVCCopyPlan` via the existing authenticated runner queue using
`materialize_pvc`, as a prerequisite before workload publication/apply for both
backends. Preserve existing fake-materialization guards. Read ADR 0013.
Derive source metadata including original UID and trusted environment class from
fresh authenticated scans; never infer nonproduction from namespace spelling.
Use out-of-band `PVCCopyPermission` and exact per-UID production approvals.
Reject all cross-tenant sources. Verify executor availability and choose the
digest-pinned image from verified compatibility manifests, not client flags.

At execution re-read UID/identity and storage, prove source writers are offline,
mount source read-only, enforce bytes/timeouts, idempotent target ownership and
safe cleanup. Authenticate evidence to the exact leased command/current
attempt/runner epoch/environment/plan digest. `pvcCopyVerified` cannot mark an
environment Ready on its own. Reject unknown JSON fields and stale evidence.
Do not add blanket approvals or perform live deployment during implementation.

Acceptance: test absent/stale UID, unknown classification, production approval
for a replaced UID, cross-tenant sources, mutable/unallowlisted images, quota
overflow, active writers, copy timeout, stale attempt/runner results and both
backend gates. Successful copy evidence unblocks only the PVC prerequisite.
