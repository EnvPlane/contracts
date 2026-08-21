# Tenant-scoped OIDC

OIDC provider records are keyed by `(tenantId, providerId)`. The API accepts
only endpoint metadata and a `SecretReference`; resolved client-secret bytes
are never persisted, returned, logged, or placed in an OAuth session.

`GET /api/v1/oidc/providers` returns redacted records for the authenticated
tenant. `PUT` and `DELETE /api/v1/oidc/providers/{providerID}` require the
tenant settings-write permission and an exact tenant context. Saving performs
HTTPS endpoint validation and issuer discovery before activation.

The existing `ENVPLANE_OIDC_*` configuration remains a read-only legacy
fallback for the default tenant only. A non-default tenant without its own
provider cannot use the default tenant's provider. OAuth state and sessions
carry the tenant ID; tenant-scoped OIDC callbacks require matching state,
nonce, and an active membership in that tenant.
