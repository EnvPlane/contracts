# EnvPilot Contracts

Shared API and domain contracts for EnvPilot services.

## Scope

- Environment, project, product, settings, bootstrap, runner, and agent domain types.
- Shared status enums and request/response payloads.
- OpenAPI specification under `openapi/openapi.json`.

## Source Origin

This repository was split from:

- `internal/domain`
- `internal/server/openapi.json`

## Notes

The module is published under `github.com/envpilot/contracts`. The domain
package still lives under `internal/domain`; a follow-up extraction will move
it to a public `domain` package and update service dependencies.
