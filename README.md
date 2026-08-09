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
import "github.com/envpilot/contracts/domain"
```

Service repositories should consume these types instead of maintaining private
copies. Update the OpenAPI document and dependent clients together with model
changes.

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
