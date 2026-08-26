# Secret materialization routes missing from canonical OpenAPI

The production control-plane registers project status, dispatch, agent claim,
and agent result routes for Secret materialization, but the canonical OpenAPI
document does not describe them. Route parity therefore fails and generated
SDK metadata cannot expose the versioned agent transport.

## Resolution

- Add all four routes to the canonical OpenAPI document.
- Describe the versioned metadata-only command and result DTOs.
- Regenerate the Go SDK contract metadata and run compatibility tests.

## Status

Resolved in the contracts release following v0.1.54.
