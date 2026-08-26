# ADR 0012: Secret materialization lifecycle

## Status

Accepted for SM-01.

## Decision

`SecretMaterializationPlan` remains contract version `v1`. Lifecycle fields are
optional on the wire so existing stored project configurations and plans remain
readable. Newly compiled plans start in `pending` with revision `1`.

Plan states are `pending`, `approved`, `materializing`, `ready`, `failed`,
`cleaning`, and `deleted`. Legal forward paths are pending to approved,
approved to materializing, materializing to ready, ready to cleaning, and
cleaning to deleted. Any active state may enter `failed` where execution can
retry only with a retryable error; terminal validation and authorization
failures require operator correction.

Item results contain only identifiers, strategy and operation names, digests,
timestamps, attempt number, status, and a redacted error code. They never carry
secret values, encrypted payloads, or generated credentials. Materialization
operations are `bind_existing_secret`, `resolve_external_secret`,
`decrypt_and_clone_secret`, `await_manual_secret_reference`, and
`generate_secret`; cleanup is always `delete_owned_secret` and is restricted to
owned targets in the exact target namespace.

Every item operation is claimed by a SHA-256 idempotency key derived from the
tenant, project, environment, template digest, target namespace, item ID, and
operation. Revision checks are optimistic: a mutation is accepted only when
the caller's expected revision equals the current revision.

## Compatibility

The canonical plan digest excludes mutable lifecycle state, revision, and item
results. This keeps the immutable desired plan identity stable while execution
progress changes. JSON encoding uses metadata-only types and omits write-only
plaintext fields.
