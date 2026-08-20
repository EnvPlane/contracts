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

The generated Go SDK embeds the SHA-256 of `openapi/openapi.json`; tests reject
stale generated output. Domain JSON round-trip tests remain the guard against
accidental field or enum changes.
