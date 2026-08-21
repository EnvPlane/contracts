# Tenant-scoped SCIM provisioning

The supported SCIM subset is tenant-scoped `Users` and `Groups`: list, PUT,
PATCH, deactivate, and group-to-role mapping. Credentials are generated once
by the tenant admin, stored only as hashes, and rotated through the protected
tenant endpoint. They are never logged or included in audit payloads.

PUT and PATCH are idempotent. User deactivation disables the matching
membership and bumps the session epoch, so existing sessions stop authenticating
without deleting the user or audit history. Lists are deterministic by resource
ID and support `startIndex`, `count`, and `userName eq/co` filters. Every
mutation emits tenant-scoped audit metadata without bearer values. Arbitrary
SCIM extensions and provisioning outside the Users/Groups subset are rejected
by the strict JSON models.
