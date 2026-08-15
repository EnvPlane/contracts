# ADR-0002: Public community implementations and future Enterprise modules

## Status

Accepted.

## Decision

EnvPlane adopts the public-core interpretation of the open-core boundary for
the code already published under Apache-2.0. Existing billing, entitlement,
identity, policy, audit, FinOps, fleet, DR and air-gap implementations in the
public control-plane repository remain part of the Community compatibility
core. Their publication rights are not revoked or retroactively restricted.

Future commercial value is delivered through managed operations, support,
separately distributed Enterprise services, and new Enterprise-only modules.
New private code may depend on versioned public contracts and extension
interfaces, but public repositories must not depend on private modules.

The chart-managed E2E fixture is test-only and remains excluded from
production builds through the `e2e` Go build tag.

## Consequences

- Existing public code remains legally and technically usable under
  Apache-2.0.
- Monetization documents must describe already-published control-plane
  capabilities as public compatibility code, not as code scheduled for
  retroactive privatization.
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

