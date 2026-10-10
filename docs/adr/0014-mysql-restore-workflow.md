# ADR 0014: Lifecycle-bound, metadata-only MySQL restoration

Status: staged for main/worker acceptance; publication remains held.
Date: 2026-10-10

## Decision

Add sibling `MySQLRestorePlan`, `MySQLRestoreSource`, `MySQLRestoreItem` and
`MySQLRestoreResult` types in `domain/mysql_restore.go`. Preserve PVC contracts
and existing materialization guards. `RunnerOperationRestoreMySQL` is
`restore_mysql`. Add optional command plan, result digest/verified metadata and
heartbeat contract version; no existing required transport fields change.

The source projection retains PVC, StatefulSet/Deployment and credential Secret
UIDs separately from sanitized manifests. It includes the selected container,
database/principal, password **key name**, pinned source image and storage and
environment classification metadata. It contains no passwords, database rows,
dumps, SQL, scripts, grants or connection strings. Unknown classification fails
closed. Cross-tenant copying is forbidden; production approval is trusted,
out-of-band and exact to namespace/PVC name/UID.

`SourceImage` must come from a trusted reviewed immutable image profile or
verified manifest, not client flags or an inferred resolution of `mysql:8.4`.
Before backup Runner must compare the pinned digest to the actual owned source
Pod/container image ID. No `SourceObservedImage` is synthesized at compilation.
Missing read/backup/lock privileges (including BACKUP_ADMIN where required by
the reviewed backup mode) are prerequisites, never grounds for automatic grants.

The item binds a new target PVC, generated Secret name/UID/key/principal, target
database and completed Secret-materialization plan digest. Target UID is required
before compilation, not discovered later and silently substituted. The existing
`SecretMaterializationItemResult` gains optional `OutputUID` (`outputUid`) for
authenticated Agent readback after materialization. Old results remain valid
without it; MySQL prerequisites do not. Main must verify current command/attempt,
managed ownership labels and plan digest before treating OutputUID as evidence.
Results are excluded from existing Secret-plan canonical digests, preserving
compatibility. Source credentials are never cloned into the target.

The plan pins helper and target MySQL images; trusted policy allowlists both.
Target storage/copy bytes and timeout use positive, aggregate, overflow-safe
bounds. Source requested bytes must match the trusted scan; target capacity must
accommodate them. `MaxBytes` caps the logical stream, not the volume allocation.
Target storage class/modes may be selected by reviewed policy for the new PVC.

## Lifecycle and validation boundary

`EnvironmentCreatedAt` is the persisted environment's lifecycle instance
timestamp, not plan compilation time. It is mandatory, canonicalized to UTC and
covered by the digest and trusted permission binding. A recreated environment
with a reused environment ID is a different instance. Main must also include
this timestamp in stable PlanID derivation and compare current CreatedAt against
the original at every prerequisite wait/publication transition. Apply the same
PlanID/wait binding to PVC plans without changing their existing wire schema.

Public entrypoints:

```go
CompileMySQLRestorePlan(MySQLRestorePlan, []MySQLRestoreSource,
    MySQLRestorePermission) (MySQLRestorePlan, error)
ValidateMySQLRestorePlan(MySQLRestorePlan, []MySQLRestoreSource,
    MySQLRestorePermission) error
(MySQLRestorePlan).Validate() error
(MySQLRestorePlan).CanonicalDigest() (string, error)
(MySQLRestoreSource).CanonicalIdentity() (string, error)
(MySQLRestoreResult).Validate(MySQLRestorePlan, commandID, attemptID string) error
```

Structural Validate is not authorization. Compilation and execution require
fresh authenticated sources and out-of-band policy, including exact generated
target Secret evidence. Strict plan decoding rejects unknown nested properties,
duplicate keys and case-insensitive aliases. No supplied digest is silently
replaced. Result evidence must match CURRENT command/attempt/plan digest and
lifecycle; it is never standalone environment Ready. Runner identity/epoch and
lease authentication remain enforced by the existing queue owner.

## Publication boundary

No publication or consumer pins until main accepts the shared schema and both
worker/API integrations, followed by an immutable freeze and green checks.
See `docs/contracts-publication-checklist.md` for workflow behavior and gates.
