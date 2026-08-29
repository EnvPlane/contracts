# Canonical OpenAPI lacks first-run and current-cluster routes

## Evidence

`control-plane/scripts/check-openapi-route-parity.sh` fails when the canonical
contracts specification is supplied. The control-plane router/spec contains
these operations that are absent from `contracts/openapi/openapi.json`:

- `GET /api/v1/first-run/progress`
- `GET /api/v1/current-cluster/preflight`
- `POST /api/v1/current-cluster/reconcile`
- `POST /api/v1/settings/authentication/first-run/transition`
- `POST /api/v1/settings/authentication/first-run/reset`

## Required fix

Add the operations and their safe schemas to the canonical contracts OpenAPI,
regenerate the SDK metadata, copy the canonical spec to control-plane, and run
the route-parity check with `ENVPLANE_CONTRACTS_OPENAPI_SPEC` set.

## Scope boundary

This is independent of the activation Settings API contract added by EP-SSO-014.
