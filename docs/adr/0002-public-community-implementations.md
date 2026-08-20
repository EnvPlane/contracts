# ADR-0002: Public community implementations and future Enterprise modules

## Status

Accepted.

## Decision

The project adopts the public-core interpretation only for repositories licensed
under Apache-2.0. `control-plane` is licensed under Business Source License
1.1 and `frontend` is proprietary; their commercial code is not part of the
Apache public compatibility core. ADR-0003 records the canonical repository
matrix and does not alter already-published Apache-2.0 rights.

Future commercial value is delivered through managed operations, support,
separately distributed Enterprise services, and new Enterprise-only modules.
New private code may depend on versioned public contracts and extension
interfaces, but public repositories must not depend on private modules.

The chart-managed E2E fixture is test-only and remains excluded from
production builds through the `e2e` Go build tag.

## Consequences

- The seven Apache-2.0 public-core repositories remain legally and technically
  usable under Apache-2.0.
- Commercial control-plane and frontend capabilities retain their declared
  licenses and must integrate with public contracts without becoming a private
  fork of the public core.
- New Enterprise features require a separate module or service boundary and
  a public-contract compatibility entry before implementation begins.
- License, provenance and contributor review remain mandatory for every new
  Enterprise module.

## Placement rule

Put domain types, OpenAPI, SDKs, cluster runtime code and security-critical
integration boundaries in public repositories. Put managed-service operations,
commercial-only policy, fleet orchestration and private deployment tooling in
separately versioned Enterprise repositories. A public package must never
import a private package or load private code dynamically.
