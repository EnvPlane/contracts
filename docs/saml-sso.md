# Tenant-scoped SAML SSO

SAML is an enterprise identity adapter separate from OIDC. The supported
policy is SP-initiated login only: the service creates a signed `RelayState`
and requires a matching `InResponseTo` in the signed response. Unsolicited
IdP-initiated responses are rejected because they have no server-created state
or tenant binding.

Provider records are keyed by `(tenantId, providerId)`. Metadata and signing
material are represented by `SecretReference` values; certificate PEM bytes
are resolved only during callback validation and are never persisted, returned,
logged, or placed in a session. The admin API does not accept a certificate
body. Signed assertions must match the configured issuer, audience, destination,
certificate validity window, assertion time window, and tenant. Replay IDs are
consumed once. Attribute mapping is limited to identity fields; group
provisioning and SCIM are not part of this flow.
