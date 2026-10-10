# ADR 0014: Lifecycle-bound, metadata-only MySQL restoration

Status: staged for main/worker acceptance; publication remains held.
Date: 2026-10-10

## Decision

Add sibling `MySQLRestorePlan`, `MySQLRestoreSource`, `MySQLRestoreItem` and
`MySQLRestoreResult` types in `domain/mysql_restore.go`. Preserve PVC contracts
and existing materialization guards. `RunnerOperationRestoreMySQL` is
`restore_mysql`. Add optional command plan, result digest/verified metadata and
heartbeat contract version; no existing required transport fields change.

MySQL heartbeat capabilities are independent of filesystem copy configuration:
`MySQLRestoreContractVersion`, `MySQLRestoreSourceNamespaces`,
`MySQLRestoreTargetImages`, `MySQLRestoreHelperImage`. Main must require an
authenticated online heartbeat no older than 90 seconds, a verified compatible
Runner image, a pinned independent helper and approved target image membership.
These are operator configuration/verified compatibility metadata, never dynamic
client flags. SQL backup scope does not inherit filesystem-offline source fences.

The source projection retains PVC, StatefulSet, Service and credential Secret
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

V1 accepts StatefulSet sources only; Deployment is explicitly unsupported.
Service name/UID/port are mandatory trusted scan metadata, so Agent UID
projection must also cover Service. `BackupAdminSecretRef` is an optional
`*MySQLRestoreCredentialRef` that becomes mandatory for `backup_lock`. Its
Namespace/Name/UID plus literal Username or UsernameKey and PasswordKey must
exactly match trusted `ApprovedBackupAdminSecrets` from normal onboarding. It
cannot alias the app credential Secret. `quiescence` instead requires a stable
QuiescenceRef in trusted `ApprovedQuiescenceRefs` and live Authority enforcement
by the executor; a job/client flag is never sufficient. Source TLS metadata
binds a same-namespace CA Secret name/UID/key and Service DNS server name.
VERIFY_IDENTITY is mandatory; no insecure option exists. Operator-provisioned
trusted TLS may be required when chart defaults are self-signed.

TLS `ServerName` is also the connection host, not merely a certificate label.
V1 permits only exact `<service>.<namespace>.svc` or
`<service>.<namespace>.svc.cluster.local`, derived from the bound source Service
and namespace. A prefix such as `<service>.<namespace>.svc.attacker.example`
does not establish Service ownership and is rejected, even if supplied in the
reviewed source projection. Main source-policy/UI validation and Worker must
enforce the same exact derivation. Custom cluster DNS requires a future trusted,
authenticated cluster-domain contract; it is not enabled by caller/profile flags.

The contract mirrors initial worker limits: matching source/target image and
database, non-root app/target principal, no system database, at most six hours
per plan, at most 1 TiB per item copy/storage, 1 KiB–16 MiB statement limit,
1–1000 tables and 1–100000000 rows. These bounds are metadata and execution
limits, not dump/row payloads.

The item binds a new target PVC, generated Secret name/UID/key/principal, target
database and completed Secret-materialization plan digest. Target UID is required
before compilation, not discovered later and silently substituted. The existing
`SecretMaterializationItemResult` gains optional `OutputUID` (`outputUid`) for
authenticated Agent readback after materialization. Old results remain valid
without it; MySQL prerequisites do not. Main must verify current command/attempt,
managed ownership labels and plan digest before treating OutputUID as evidence.
Results are excluded from existing Secret-plan canonical digests, preserving
compatibility. Source credentials are never cloned into the target.

`TargetRootPasswordKey` is mandatory and covered by the plan digest. It must be
a valid Secret key different from `TargetPasswordKey`; trusted target-Secret
approval additionally binds `RootPasswordKey`, preventing a re-sealed plan from
selecting a different key. The trusted `mysql-password-v1` generator independently
produces root and app credentials; its fixed root key is `MYSQL_ROOT_PASSWORD`
(`MySQLRestoreRootPasswordKey`). Main/runtime accept that generator constant,
never a client-controlled key flag. Metadata key separation does not prove value
independence: generator and executor tests must verify distinct generated values
without persisting/logging them in contracts. Root init must not reuse app values.

### Target layout and sealed workload publication invariant

`MySQLRestoreTargetDataSubPath = "mysql"` is the shared fixed target layout.
Worker target initialization writes/restores data in that subdirectory, and the
generated **feature** SQL workload must mount its new target PVC with
`subPath: mysql`. Existing source workloads may mount their PVC root; never
rewrite source mounts or mount source data in restore helpers.

Before sealing the feature template revision, main must bind the target data
subpath, reviewed pinned MySQL image (`SourceImage`/matching target image),
generated app/root Secret refs and independent root-password key together in
the desired workload metadata. The revision digest must cover those rendered
bindings. Both backend publication gates must preserve them and use the same
target PVC as the verified restore. Publishing a root mount could start a new
empty database beside the restored data; refuse any layout/image/Secret drift
rather than releasing workload on copy evidence alone.

No extra plan wire field is added: v1 has one fixed target subpath, not a
client-selectable path. Main/worker tests must verify the fixed layout and sealed
revision bindings; contracts constant tests alone are not runtime/live proof.

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
