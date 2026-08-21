# Versioned plan catalog

The canonical plan catalog is defined by `contracts/domain/plan.go` and is
validated before it is consumed by entitlement resolution. The catalog is
descriptive only: EP-COM-013 does not enforce limits, persist subscriptions,
or make billing decisions.

## Versioning

Every plan has both `schemaVersion` and `effectiveVersion`. A released plan
revision is immutable. A future change creates a new effective version in
source control; it does not mutate an entitlement snapshot that already stores
the previous `planVersion`. The resolver keeps one active definition per plan
ID, while historical revisions remain represented by immutable source history
and the snapshots that reference them.

## Free bundle

The `free` plan is the Community default bundle:

| Key | Limit |
| --- | ---: |
| `projects.max` | 3 |
| `clusters.managed.max` | 1 |
| `environments.active.max` | 2 |
| `environment.ttl.max_hours` | 72 |
| `environment.pin.max_hours` | 168 (7 days) |
| `users.operators.max` | 3 |
| `audit.retention_days` | 7 |

It exposes OIDC, Helm Direct, GitOps Flux, projects, environments, and audit
as catalog capabilities. SAML, SCIM, granular RBAC, Argo, custom policy,
FinOps allocation, upgrade waves, and SLA support are disabled in this bundle.
These values describe the bundle; enforcement remains outside this ticket.

## Audit retention

Audit retention is evaluated per tenant from the immutable entitlement snapshot
using UTC. Free retains seven days; a paid plan uses its
`audit.retention_days` limit. Audit reads apply the same lower time boundary as
purge, so API and UI cannot read records outside the tenant entitlement.
Purge runs in bounded tenant-scoped batches. A record marked `legalHold` is
never eligible for deletion. SIEM export is a separate opt-in capability,
`audit.siem_export`; it is enabled only by entitlement and an explicit audit
export permission. The exporter sends redacted metadata only and never blocks
primary audit writes.

## Keys and compatibility

Canonical feature and limit keys are the dotted identifiers exported from the
domain package. Existing short feature names and `max*` limit names remain
explicitly allowlisted as compatibility aliases for current callers. Unknown
keys are rejected. JSON maps cannot contain duplicate keys by representation;
duplicate plan IDs are rejected as well.

## Policy evaluation

Policy bundles are versioned and deterministic. Built-in declarative rules run
after authentication and quota checks and before resource side effects for
environment creation and cluster onboarding. Denials include stable rule IDs,
the bundle version, and a safe explanation; denials are audit-recorded without
request payloads. Create evaluation failures are fail-closed. Cleanup/recovery
may fail open with a warning so cleanup cannot strand resources.
