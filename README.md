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

The current Go package path is still `envpilot/internal/domain` from the monorepo. A follow-up migration should publish these contracts as a stable public module path and update service imports.
