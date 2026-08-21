# ADR-0004: AI architecture and versioned contracts

## Status

Accepted for EP-AI-001. The requested `ai-integration-tickets/README.md` was not
present in the workspace when this decision was written; this ADR follows the
ticket scope and existing contracts repository conventions.

## Decision

AI is a control-plane-only capability. The control plane may prepare a typed request
from tenant-scoped, allowlisted evidence references and may return a typed diagnosis
result. AI providers never receive direct Agent, Runner, Kubernetes, database, or
queue access. AI has no action capability: it cannot create, mutate, delete, repair,
approve, execute commands, or change deployment state.

All public AI domain objects use `schemaVersion: "1"` and live in
`contracts/domain/ai.go`. They carry tenant and request/run identity so results cannot
be detached from the authorization context that produced them.

## Data flow and trust boundaries

1. An authenticated control-plane request establishes the tenant and subject scope.
2. The control plane selects allowlisted source records and emits only
   `AIEvidenceReference` values containing source type, stable source ID, tenant ID,
   optional observation time, and optional digest. Raw payloads, manifests, logs with
   secrets, tokens, kubeconfigs, and provider credentials are outside the contract.
3. A future provider adapter receives the versioned request through a control-plane
   port. The adapter is not linked to Agent/Runner/Kubernetes clients.
4. The control plane records `queued`, `running`, `succeeded`, `failed`, or `canceled`
   status and validates the result before exposing it to callers.
5. The result is advisory only. Any human or existing policy/action path must make a
   separate authorized decision.

Tenant checks happen before evidence selection and again when validating returned
evidence. A source ID is not sufficient authority by itself.

## Contract rules

- A run status has exactly one of `queued`, `running`, `succeeded`, `failed`, or
  `canceled`.
- Evidence references use a stable source type and ID and never embed arbitrary raw
  payloads.
- A `diagnosis` result must contain at least one same-tenant evidence reference and a
  known confidence level. Confidence cannot compensate for missing evidence.
- `insufficient_evidence` must carry zero evidence, zero score, and `unknown`
  confidence. It is a successful, explicit inability-to-diagnose outcome, not a
  fabricated low-confidence diagnosis.
- Provider failures are reduced to the allowlisted classes in the domain contract;
  provider response bodies, secrets, and credentials are not propagated.

## Threat model and failure policy

| Threat | Boundary/control | Policy |
|---|---|---|
| Secret leakage | typed references only, redaction before provider call, no raw evidence field | reject unsafe evidence and audit a safe classification |
| Prompt injection | source content is untrusted evidence, not instructions or authority | treat provider output as advisory and require normal authorization for actions |
| Cross-tenant access | tenant ID is bound to request, evidence, and result | reject mismatched evidence/results; never fall back to global reads |
| Hallucination | evidence count and outcome validation | expose `insufficient_evidence`; do not present unsupported diagnosis as fact |
| Provider outage/timeout | normalized provider error classes and bounded adapter retries in a later ticket | fail closed for AI-dependent recommendations; preserve existing non-AI operations |
| Unbounded cost | future adapter port must enforce request budgets, timeouts, and rate limits | no provider call or billing integration in EP-AI-001 |

Provider errors are observable through safe class/code fields only. A provider failure
must not block read, delete, cleanup, export, or existing runtime authentication paths.

## Non-goals

- No external SDK, provider credentials, network call, prompt template, or provider
  selection logic.
- No AI API route, UI, persistence migration, background worker, or action execution.
- No direct integration with Agent, Runner, Kubernetes, SCM, or deployment backends.
- No claim that confidence is calibrated probability; calibration and evaluation are
  later work.
