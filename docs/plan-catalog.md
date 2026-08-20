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

## Keys and compatibility

Canonical feature and limit keys are the dotted identifiers exported from the
domain package. Existing short feature names and `max*` limit names remain
explicitly allowlisted as compatibility aliases for current callers. Unknown
keys are rejected. JSON maps cannot contain duplicate keys by representation;
duplicate plan IDs are rejected as well.
