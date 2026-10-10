# ADR 0013: Metadata-only, identity-bound PVC copy workflow

**Status:** Accepted for contracts; executor integration is separate.
**Date:** 2026-10-10
**Deciders:** EnvPlane workflow owners

## Context

`StatefulExecutionPlan` expresses metadata strategies but is not evidence of a
real copy. Sanitized manifests can remove `metadata.uid`. Copying a name without
its scanned UID can read a replacement PVC, and copy flags or a self-recomputed
digest cannot establish source ownership, production approval or readiness.

## Decision

Add independent `PVCCopyPlan`/`PVCCopyItem` contracts in `domain/pvc_copy.go`.
Only `filesystem_offline` and filesystem volumes are supported in v1. Plans
contain no credentials, rows, dumps, scripts, commands, arbitrary manifests or
data payloads. Names are validated as Kubernetes identifiers. Opaque binding
IDs support UUIDs, digests and slash-delimited IDs (up to 512 bytes), but must
never be used as file paths or interpolated shell arguments.

Public entrypoints:

```go
CompilePVCCopyPlan(plan PVCCopyPlan, sources []PVCCopySource,
    permission PVCCopyPermission) (PVCCopyPlan, error)
ValidatePVCCopyPlan(plan PVCCopyPlan, sources []PVCCopySource,
    permission PVCCopyPermission) error
(PVCCopyPlan).Validate() error
(PVCCopyPlan).CanonicalDigest() (string, error)
(PVCCopySource).CanonicalIdentity() (string, error)
```

The compiler fills only an empty contract version and empty source identity
digests. Selection, UID, method, byte sizes, timeouts and tenant are explicit.
It deep-copies slices and rejects any supplied mismatched digest. Canonical
digests use JSON with the plan digest blanked, sorting items by unique ID and
namespace/access-mode sets. Source identity covers UID, tenant, namespace/name,
storage class, modes, requested bytes, volume mode and environment class.
Unknown fields and duplicate keys in plan JSON are rejected even when nested
inside a normally decoded `RunnerCommand`.

The server projects `PVCCopySource` from an authenticated fresh scan. Retain the
original UID in additive `ResourceSnapshot.SourceUID` (`sourceUid`), independently
of manifest sanitization. Missing UID, missing/unknown environment classification,
nonpositive sizes and block volumes fail closed. Kubernetes' absent volumeMode
may be explicitly projected as `Filesystem` by the scanner. Classification must
come from trusted source Namespace/PVC `envplane.io/environment-class` labels or
policy, never from a user-chosen namespace name. Supported classifications are
`production`, `development`, `staging`, `test`, `preview`. Treat absent/unknown
classification as production in trusted projection only if production policy
allows it; otherwise reject. No live label mutation is required or authorized.

`PVCCopyPermission` is trusted out-of-band context, not a transport field. It
binds exact plan/tenant/project/environment/revision/template digest/target and
source namespace set, permits exact digest-pinned image references, and caps
storage/copy bytes and timeout. Derive allowed images from the verified compatible
runner manifest, not arbitrary server narrative or client flags. Cross-tenant
sources are always rejected in v1, even with production approval. Production
approval is an exact namespace/name/UID entry in `ApprovedProductionSources`.
Source and target namespaces cannot match. Source selection, item IDs and target
names must be unique. Storage class and modes must match the trusted source;
read-only target modes are unsupported. Target requested bytes and item maxBytes
must accommodate scanned requested bytes; total requested bytes and maxBytes
are checked independently with overflow-safe quota arithmetic. Timeout is 1 to
86400 seconds, bounded by both plan and trusted policy. Sizes are exact integer
bytes; quantity parsing and freshness remain the server's responsibility.

`Validate()` checks structure and integrity only. A recomputed digest is not
authorization: use `ValidatePVCCopyPlan` with fresh trusted sources and policy at
compilation and before execution. A consumer that has only the plan must receive
equivalent trusted immutable source/policy bindings through the authenticated
queue and verify them; never treat successful structural validation as approval.

Additive transport:

- `RunnerCommand.PVCCopyPlan *PVCCopyPlan` (`pvcCopyPlan,omitempty`).
- `RunnerCommandResult.PVCCopyPlanDigest string` and `PVCCopyVerified bool`.
- `RunnerHeartbeatRequest.PVCCopyContractVersion string`.
- `RunnerOperationMaterializePVC = "materialize_pvc"`.

Before compiling, the parent verifies authenticated executor capability and
compatible artifact provenance. Heartbeat version alone is not proof of a
verified executor. Both deployment backends must enqueue the new operation
before workload apply/publication. Result digest/verified metadata is valid only
for the exact leased command, current attempt, runner epoch, environment and
plan digest. It satisfies a PVC prerequisite, never standalone environment
Ready. Preserve the existing `StatefulExecutionPlan` API and its fake/materialized
guard unchanged. Executors must re-read source UID and identity, enforce offline
writers, bounds and image provenance, and prevent source deletion/mutation.

## Options considered and trade-offs

Extending `StatefulExecutionPlan` would conflate descriptive strategies with
execution proof and risk weakening compatibility guards. An independent plan
keeps that boundary explicit, at the cost of a separate lifecycle prerequisite.
Allowing snapshots/online copies or cross-tenant exceptions would require new
consistency and authorization contracts; they are intentionally deferred.

## Consequences and action items

Existing consumers retain their APIs and omit new optional metadata. Update the
canonical OpenAPI and regenerate SDK metadata when these schemas change.
Parent/worker integrations must implement scan projection, capability gating,
offline enforcement, command/attempt-bound evidence and workload publication
gates separately. See `issues/2026-10-10/pvc-copy-integration/001-executor-boundary.md`.
