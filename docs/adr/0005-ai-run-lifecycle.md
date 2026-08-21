# ADR-0005: Durable AI run lifecycle

## Status

Accepted for EP-AI-004.

## Decision

AI runs are durable control-plane metadata records scoped by `tenantId` and
`projectId`. The record stores lifecycle state, idempotency key, provider/model
identifiers, prompt-template version, context hash, bounded token counters,
latency and a sanitized provider error category. Raw prompts, contexts and
provider responses are not persisted by default.

Every caller supplies an already authenticated tenant/project permission
decision to the lifecycle service. Storage keys and PostgreSQL RLS independently
enforce the tenant boundary. A cross-tenant lookup is indistinguishable from
not-found at the store boundary.

Valid transitions are `queued -> running -> succeeded|failed|canceled` and
`queued -> failed|canceled`. Terminal states cannot be changed. Idempotency is
unique per tenant, project and key; a retry returns the original run without a
second provider call.

The service has bounded concurrent, request, input-token, output-token and
duration limits. Reconciliation marks stale running records as failed with the
sanitized `timeout` category. Retention deletes only expired `AIRun` rows in a
bounded batch; it does not touch environments, jobs, audit records or other
operational state. Metrics, when wired, must use fixed result/status categories,
never tenant or project IDs as labels.

Production activation remains disabled until the commercialization entitlement
gate is explicitly connected. The development JSON store remains available for
the existing local persistence mode.

## Non-goals

- No diagnosis endpoint, UI, background provider worker or automatic action.
- No raw prompt/context/provider response storage.
- No subscription or entitlement implementation in this ticket.
