# ADR-0001: Open-core and Enterprise boundary

## Status

Accepted.

## Context

EnvPlane is a Kubernetes environment control plane. The cluster-side runtime
must remain inspectable by customers, while identity, compliance and managed
service operations provide commercial Enterprise value. All current component
repositories are Apache-2.0 and already-published rights cannot be revoked.

## Decision

The following repositories form the public compatibility core:

| Repository | Public responsibility |
|---|---|
| `contracts` | domain types, OpenAPI and SDK contracts |
| `agent` | inspectable cluster observer |
| `runner` | inspectable constrained execution runtime |
| `webhook` | SCM webhook validation and normalization |
| `bootstrap` | onboarding/discovery core |
| `gitops` | rendering and repository-writing core |
| `control-plane` | Community lifecycle control-plane core |
| `frontend` | Community UI |
| `deploy` | Community installation and signed release metadata |

Private Enterprise modules provide entitlement/billing, managed SaaS
operations, SAML/SCIM, advanced RBAC, policy/compliance/SIEM, FinOps, fleet
operations, HA/DR and air-gapped packaging.

Private modules may depend only on versioned public contracts and declared
extension interfaces. Public modules must not import private modules, fetch
private code at runtime or contain hidden license bypasses. Community builds
provide a complete Free flow through explicit community implementations; an
Enterprise module failure must not block read, delete, cleanup or export.

Release artifacts compose a public compatibility set first. An Enterprise
artifact adds private modules as separately versioned services or plugins and
records their compatibility against the public set. The control plane remains
the only authority that calls extension interfaces.

## Extension interfaces

The stable interface families are entitlement, identity, policy evaluation,
audit sink and FinOps. They use public contracts, receive a tenant-scoped
principal where relevant, return typed data only, and must not receive raw
provider credentials except through a dedicated secret-reference resolver.

## Legal and release checklist

Before changing repository visibility, license text or artifact distribution:

1. inventory source provenance and third-party notices;
2. confirm Apache-2.0 obligations for existing releases;
3. approve a Contributor License Agreement and trademark policy;
4. verify that Enterprise code is independently authored and separated;
5. publish upgrade, rollback and support boundaries;
6. perform legal review and record the decision.

## Consequences

The product can offer a useful Community edition without exposing commercial
service operations. This decision does not change current licenses or GitHub
visibility; those are explicit legal and business actions.
