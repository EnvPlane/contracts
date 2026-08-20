# ADR-0003: Open-core repository and release policy

## Status

Accepted. This ADR refines ADR-0001 and ADR-0002 without changing licenses,
repository visibility, or already-published rights.

## Decision

The nine current repositories are the public EnvPlane compatibility core:

| Repository | Public responsibility |
|---|---|
| `contracts` | Versioned domain, OpenAPI, SDK and compatibility contracts |
| `agent` | Cluster discovery and observed inventory |
| `runner` | Constrained runtime transport and execution |
| `webhook` | SCM webhook validation and normalization |
| `bootstrap` | Project onboarding and discovery |
| `gitops` | Rendering and repository-writing integrations |
| `control-plane` | Community lifecycle and orchestration core |
| `frontend` | Community web UI |
| `deploy` | Charts, installation and signed release metadata |

Future private Enterprise modules are separate implementation boundaries for
managed operations, commercial policy and entitlements, fleet operations, and
private deployment or air-gapped packaging. These are logical boundaries, not
new repositories or visibility changes made by this ADR.

Private Enterprise code may depend on versioned public contracts and declared
extension interfaces only. Public code must not import a private module, use a
private fork of a public repository, fetch private code dynamically, or make a
private service mandatory for the Community flow. Public interfaces are the
only dependency direction across the boundary.

## Extension interfaces

Enterprise integrations use typed, versioned interfaces for identity,
entitlements, policy evaluation, audit sinks, FinOps and managed operations.
Each interface receives a tenant-scoped principal and typed references, never
plaintext credentials. Enterprise failures must degrade according to the
public contract and must not bypass tenant isolation or ownership checks.

## Release artifact composition

Every release is composed from the public compatibility set first: contracts,
public service/runtime images, frontend, charts and signed compatibility
metadata. Enterprise artifacts add separately versioned private services or
plugins, declare their public-contract compatibility, and never replace the
public artifact with a private fork. Public release metadata must not contain
private source, credentials or private implementation details.

## Legal checkpoint

Before any change to licensing, repository visibility, artifact distribution or
private module publication, obtain explicit legal and business approval for:

1. Apache-2.0 notices, attribution and third-party provenance;
2. CLA scope, contributor provenance and copyright ownership;
3. EnvPlane trademark usage, naming and redistribution policy;
4. separation and independent authorship of Enterprise code;
5. public/private upgrade, support, rollback and security boundaries;
6. release artifact contents and customer license terms.

Until all six checkpoints are recorded, no automated change to license,
visibility or publication policy is permitted.

## Consequences

The public repositories remain useful and independently buildable. Enterprise
features can evolve privately behind stable contracts, while the dependency
direction, release composition and legal decision points remain auditable.
