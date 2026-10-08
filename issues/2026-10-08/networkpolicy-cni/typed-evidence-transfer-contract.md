# Typed network policy evidence transfer

Status: implemented locally; requires publication before independent release CI.

Domain types and canonical OpenAPI define bounded safe report, challenge,
submission and observation. Only explicit passed ingress+egress with cleanup can
be passed; reason/scope enums exclude arbitrary output. Observation freshness is
server-derived, never client-controlled. Runtime bearer and nonce are submission
only and are absent from the public observation. Cluster UID is pinned evidence
from an authenticated Agent, not independent attestation.

Codex review prompt: preserve additive compatibility and tenant/auth boundaries;
update all direct consumer pins after publishing, regenerate SDK/OpenAPI copies,
reject contradictory/unknown-scope passed evidence, and keep stale observations
non-authoritative. Unit tests and generated metadata checks run locally.
