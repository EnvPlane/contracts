# ADR-0003: Open-core repository and release policy

## Status

Accepted. This ADR refines ADR-0001 and ADR-0002 without changing licenses,
repository visibility, or already-published rights.

## Decision

The Apache-2.0 public compatibility core has seven repositories:

| Repository | Public responsibility |
|---|---|
| `contracts` | Versioned domain, OpenAPI, SDK and compatibility contracts |
| `agent` | Cluster discovery and observed inventory |
| `runner` | Constrained runtime transport and execution |
| `webhook` | SCM webhook validation and normalization |
| `bootstrap` | Project onboarding and discovery |
| `gitops` | Rendering and repository-writing integrations |
| `deploy` | Charts, installation and signed release metadata |

The remaining current repositories are commercial code outside that Apache
core: `control-plane` is licensed under Business Source License 1.1 and
`frontend` under its proprietary license. They may consume and implement
public contracts but must not become private forks of the Apache core.

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

The seven Apache repositories remain useful and independently buildable.
Commercial repositories and future Enterprise features can evolve behind stable
contracts, while the dependency direction, release composition and legal
decision points remain auditable.
