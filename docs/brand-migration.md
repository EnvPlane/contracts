# EP-BRAND-006: shared contract migration

The contracts module keeps its established import path
`github.com/envplane/contracts` during the migration window. Renaming the Go
module or package paths in place would break every component, so a future major
version must provide an explicit compatibility module before removal.

The v1 JSON wire contract is also stable: API paths, JSON field names, enum
values, authentication schemes and generated SDK symbols are not renamed for
branding. OpenAPI now records EnvPlane as the canonical product and exposes
`x-branding` deprecation metadata. `ENVPLANE_*` names in security descriptions
are documented as legacy aliases, not new wire fields.

Sequencing:

1. Current releases: use EnvPlane display text and canonical `ENVPLANE_*`
   documentation while consumers continue importing the existing module and
   reading legacy identifiers.
2. Next minor releases: publish generated SDKs and compatibility manifests from
   the canonical OpenAPI document; add dual-read aliases in each runtime.
3. Future major release: only after downstream repositories have migrated, add
   a versioned module/API schema and announce removal of legacy machine names.

EP-COM-039 starts the low-risk environment-variable slice. The control plane
reads these aliases with canonical-over-legacy precedence:

| Canonical | Legacy | Scope |
| --- | --- | --- |
| `ENVPLANE_ADDR` | `ENVPILOT_ADDR` | listen address |
| `ENVPLANE_DATA_DIR` | `ENVPILOT_DATA_DIR` | local data path |
| `ENVPLANE_GITOPS_DIR` | `ENVPILOT_GITOPS_DIR` | local GitOps path |
| `ENVPLANE_DOMAIN_ROOT` | `ENVPILOT_DOMAIN_ROOT` | preview domain root |
| `ENVPLANE_METRICS_ADDR` | `ENVPILOT_METRICS_ADDR` | metrics listener |
| `ENVPLANE_DEPLOYMENT_BACKEND` | `ENVPILOT_DEPLOYMENT_BACKEND` | backend selector |

The Helm chart emits only the canonical names. When a legacy name is read, the
control plane records a name-only diagnostic and increments
`envplane_legacy_environment_variable_uses_total`. If both names are set, the
diagnostic records only the two names and the fixed canonical precedence; no
environment value is logged or exported. Credential-bearing variables and
module, OCI, Kubernetes, persisted, and queue identifiers remain unchanged.

The generated Go SDK embeds the SHA-256 of `openapi/openapi.json`; tests reject
stale generated output. Domain JSON round-trip tests remain the guard against
accidental field or enum changes.
## Versioned legacy machine inventory

`docs/legacy-machine-identifiers.json` is the canonical, versioned inventory of
legacy `envplane` and `ENVPLANE` machine identifiers. It records environment
variables, Go modules, OCI and Helm artifacts, Kubernetes names and labels, CLI
and webhook commands, durable persistence/metrics/queues, and external URLs.
Each entry records its owning repository, read/write path, compatibility risk,
dual-read migration strategy, and earliest removal major.

New code must dual-read old identifiers before introducing an alias, emit only
safe canonical values for new resources, and preserve idempotent ownership and
tenant boundaries. The inventory validator rejects unknown categories and can
scan a repository with `scripts/check-legacy-machine-inventory.sh --scan-root`
to prevent unclassified machine identifiers. No identifier is renamed as part
of inventory work; removal requires a future major release and migration proof.
