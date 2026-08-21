# Quota decision contract

`internal/quota.Decide` is side-effect-free. It receives the trusted tenant
context, feature and limit keys, current usage, requested delta, and the
resolved immutable entitlement snapshot. It never reserves usage and never
reads billing credentials.

## HTTP semantics

- A disabled feature returns HTTP `403` with `code=quota_feature_disabled`.
- A bounded limit exceeded returns HTTP `429` with `code=quota_exhausted`.
- `current + requestedDelta == limit` is allowed; values above the limit are
  rejected. Arithmetic overflow is treated as exhausted, never as allowed.
- A missing limit key is explicitly unlimited.

The Free bundle enforces 3 operator memberships, 3 projects, 1 managed
cluster, 2 active environments, 72 hours maximum TTL, and 168 hours (7 days)
maximum pin duration. Reservations are tenant-scoped and must be released on
failed writes; deletion and read paths do not consume quota.

The canonical `QuotaError` fields are `code`, `feature`, `limit`, `current`,
`requested`, `plan`, `upgrade_url`, and `request_id`. None may contain payment
instrument data or provider credentials. The legacy generic `Error` response
and `limit_value` field remain available for existing v1 responses.
