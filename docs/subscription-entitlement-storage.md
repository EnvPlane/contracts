# Subscription and entitlement storage

Subscriptions and materialized entitlements are tenant-scoped records. A
subscription stores only provider identity and references (`provider`,
`providerRef`, and `providerEventId`), lifecycle state, billing period, grace
period, plan/version, and source. Card numbers, payment tokens, and other
payment instrument data are never stored.

The default tenant receives an idempotent `free` subscription through the
PostgreSQL migration and JSON store bootstrap. Subscription state transitions
are validated by the domain contract and by the database check constraint.
Provider event IDs are unique for non-empty provider/event pairs; empty event
IDs are allowed for locally-created subscriptions.

Entitlements are materialized per tenant with plan/version, copied feature and
limit maps, source subscription ID, and update time. Reads and writes require
the tenant key in both JSON and SQL stores. This ticket adds persistence only;
quota evaluation and subscription enforcement remain outside its scope.

The control-plane `EntitlementResolver` is the only runtime composition path:
plan defaults are applied first, then licensed overrides, then tenant-specific
overrides. Expired overlays are ignored, snapshots are defensively copied,
revisions hash only the effective state, and the bounded cache is invalidated
per tenant or globally. Resolver input must carry the same tenant identity as
the request; it is never inferred from an untrusted override.
