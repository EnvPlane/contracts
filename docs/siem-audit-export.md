# SIEM audit export

SIEM export is an opt-in extension of the audit ledger. The effective
entitlement must enable `audit.siem_export`, and the caller must have the
`audit.export` permission. The primary audit append is independent from sink
availability and never waits for network delivery.

The exporter accepts only `SIEMAuditEvent` metadata. Payloads, repository URLs,
branches, manifests, credentials, tokens, and user email are not part of the
event contract. Events are keyed by tenant and stable event ID; a per-tenant
cursor suppresses already acknowledged cursors while retaining at-least-once
delivery semantics when sink or cursor persistence fails.

The queue is bounded. Failed sends retry a bounded number of times and then
become tenant-filtered dead letters. HTTPS requires TLS except loopback test
endpoints. Remote syslog requires TCP over TLS; plain TCP is allowed only for
loopback development endpoints. Tenant queues are flushed independently, so a
sink configured for one tenant cannot receive another tenant's events.
