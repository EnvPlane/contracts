# EnvPlane Contracts

Shared domain and API contracts for the [EnvPlane](https://envplane.dev)
platform. This repository is the compatibility boundary between the control
plane, frontend, agent, runner, webhook, and deployment components.

## Contents

- Environment, project, product, settings, and remote-cluster models.
- Bootstrap sessions and lifecycle status enums.
- Shared request and response types.
- Canonical OpenAPI specification at `openapi/openapi.json`.

## Consumers

The Go module keeps its established path for compatibility:

```go
import "github.com/envplane/contracts/domain"
```

Service repositories should consume these types instead of maintaining private
copies. Update the OpenAPI document and dependent clients together with model
changes.

Legacy machine identifiers are catalogued in
[`docs/legacy-machine-identifiers.json`](docs/legacy-machine-identifiers.json).
They are compatibility contracts rather than product branding and must follow
the documented migration policy.

The versioned branding and wire-compatibility sequencing is documented in
[`docs/brand-migration.md`](docs/brand-migration.md). JSON fields, enum values,
module paths and generated SDK symbols are not renamed in place.

The public and Enterprise module boundary is recorded in
[`docs/adr/0001-open-core-and-enterprise-boundary.md`](docs/adr/0001-open-core-and-enterprise-boundary.md).
Stable extension interfaces for entitlement, identity, policy, audit and FinOps
are defined in [`domain/extensions.go`](domain/extensions.go). They accept only
tenant-scoped typed data; Enterprise implementations remain out of this module.

## Development

```bash
go test ./...
go vet ./...
```

## Related components

- [Control Plane](https://github.com/EnvPlane/control-plane)
- [Frontend](https://github.com/EnvPlane/frontend)
- [Agent](https://github.com/EnvPlane/agent)
- [Runner](https://github.com/EnvPlane/runner)

## Status

Private product contract under active development.
