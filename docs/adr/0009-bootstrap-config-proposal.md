# ADR-0009: Typed Bootstrap configuration proposal

## Status

Accepted for EP-AI-014.

## Decision

Bootstrap proposals are typed previews over the existing wizard schema. The
server resolves component references and GitOps targets from the authorized
project catalog. A proposal is never persisted or applied automatically.

The shared `BootstrapConfigProposalFields.Validate` contract is used by both
manual update projection and proposal generation. Existing manual SCM, chart,
policy, and lifecycle validation remains the final authority before any
session mutation.

## Non-goals

- Provider/model calls, proposal persistence, approval, or execution.
- Arbitrary YAML, templates, commands, namespaces, credentials, or secret data.
