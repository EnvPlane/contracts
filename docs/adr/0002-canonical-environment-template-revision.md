# ADR-0002: Canonical environment template revisions

## Status
Accepted for EP-TPL-001.

## Decision
The canonical lifecycle is `ResourceScanReport -> EnvironmentTemplateRevision -> EnvironmentReleasePlan`. A revision is an immutable, tenant/project/cluster-bound snapshot with a SHA-256 digest. Environments pin the revision and render-input digest, so later project edits cannot mutate an existing environment.

Helm Direct and FluxCD consume the same release-plan contract. They must not replace generated resources with a fixture chart or reinterpret discovery. Secret bytes, credentials and kubeconfig data are not contract fields; only non-sensitive references may be persisted.

Legacy compiled configurations remain readable. New compilation must bind a revision ID/digest; an unbound runtime caller must return `template_revision_required` rather than silently using `manifestTemplates` or a fixture chart. Scanner and runtime apply remain separate tickets.

## Consequences
Revisions and plans use tenant-scoped immutable storage with database row-level isolation and digest validation.
