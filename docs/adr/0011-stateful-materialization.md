# ADR 0011: Stateful dependency materialization

## Status

Accepted for AI-030.

## Decision

Secret, PVC, and database materialization is represented by metadata-only
`SecretMaterializationPlan` and `StatefulExecutionPlan` contracts. Plans carry
stable references, target ownership, bounded retention, masking policy refs,
encryption-key refs, and generated-credential rotation policy. They never carry
secret plaintext, database rows, dumps, kubeconfig material, or generated
credential values.

Every stateful dependency has an explicit strategy: reference, generate, seed,
encrypted clone, snapshot clone, database restore, or isolated external service.
Production sources require a masking policy and explicit approval. Cross-tenant
sources require explicit approval and are rejected when the source namespace is
not allowlisted. All owned targets are restricted to the feature namespace;
source records are never cleanup eligible. Every execution step has a stable
idempotency key and bounded retry/retention policy.

The agent only compiles and validates this plan. Typed platform executors resolve
references and perform encryption, masking, credential generation, clone, seed,
restore, binding, readiness, and cleanup operations. They must claim steps by
the idempotency key and must not accept model narrative or raw data as input.
