# Signed offline Enterprise license

`LicenseGrant` is versioned and tenant-bound. It contains customer and plan
metadata, limits, issue/expiry timestamps, and a stable license ID. The
signature covers the canonical JSON encoding of the grant. The control plane
accepts legacy Ed25519 verification-key configuration and versioned public-key
entries for Ed25519 or ECDSA P-256/P-384/P-521. Private signing keys are never
part of EnvPlane artifacts or runtime configuration.

The verifier checks schema, required identifiers, issue/expiry ordering,
signature, trusted key ID, tenant binding, and current time. Key rotation is
represented by multiple configured key IDs; an installed grant keeps its key
ID and can be replaced by a newly signed grant without changing tenant state.

The local license store persists the signed grant and the last valid grant with
tenant-keyed records. After expiry, the resolver may use only the cached
last-valid grant during the configured bounded grace period. After grace, paid
entitlements stop; read, delete, cleanup, and export operations remain
available through the existing downgrade behavior.

Operators can use:

```bash
envplane license install --api https://control-plane.example --tenant TENANT_ID --path signed-license.json
envplane license inspect --api https://control-plane.example --tenant TENANT_ID
```

The CLI sends the signed document to the tenant-scoped API and prints only the
verified grant metadata. It never prints or accepts a private signing key.
