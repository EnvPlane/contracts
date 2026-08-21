# ADR-0007: Deterministic AI FinOps explanation

## Status

Accepted for EP-AI-012.

## Decision

FinOps explanation is a read-only projection of measured allocation, budget,
forecast, and anomaly contracts. The control plane remains authoritative for
all arithmetic and decisions. A model is not used to calculate cost, forecast,
currency conversion, anomaly thresholds, quota outcomes, or lifecycle actions.

Unknown price and partial data remain explicit in the response. Numeric fields
are marked `known: false` when their source values are not complete. Evidence
references contain stable source IDs and tenant scope, never raw resource
payloads or provider responses.

Typed recommendations are configuration proposal fields only. They require the
existing preview and explicit approved-action flow; no recommendation is
automatically applied, and no environment is deleted or reduced.

## Non-goals

- Cloud invoice or savings claims without provider-backed measurements.
- Notifications, automatic quota/policy decisions, or lifecycle mutation.
- Arbitrary natural-language configuration or direct provider calls.
