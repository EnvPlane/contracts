# ADR 0010: Environment template synthesis

## Status

Accepted for the AI-029 contract and agent slice.

## Decision

Template synthesis consumes only a tenant-scoped, sanitized `ResourceSnapshot` set and its validated `ServiceGraph`. It does not query Kubernetes, SCM, GitOps, or secret storage. The output is the versioned `EnvironmentTemplateSynthesis` contract containing an immutable `EnvironmentTemplateRevision`, provenance-bearing decisions, unresolved dependency issues, and a digest.

Built-in portable resources are parameterized through typed inputs. Runtime children are ignored. Shared or foreign resources are references requiring review. Secrets, persistent storage, unsupported kinds, ambiguous graph findings, and external endpoints are blocked until an operator selects an explicit strategy. No raw Secret data or unrestricted file content is accepted by the contract.

The synthesis result is deterministic: source namespaces, resources, decisions, and issues are sorted before revision and result digests are calculated. A result is not autonomously applicable when any unresolved or review-required decision remains. Rendering continues through the existing deterministic environment renderer; synthesis does not introduce a second write path.

## Boundaries and follow-up

The control plane may persist or expose this review artifact through the existing template APIs, while the frontend and GitOps layers remain consumers of the versioned revision and renderer. Approval, execution, rollback, and cluster/SCM mutations remain outside this slice and must use the existing agent orchestrator and typed tool gateway.

Golden end-to-end coverage must compare the sanitized base inventory with the rendered feature inventory, routing, configuration, health, and cleanup for one and multiple source namespaces. Cases with ambiguous selectors, external dependencies, unsupported CRDs, secrets, or PVCs must assert a blocked autonomous decision.
