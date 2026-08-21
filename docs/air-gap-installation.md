# Offline Enterprise installation runbook

## Export

1. Select the exact EnvPlane image and chart set from the published
   compatibility manifest. Every artifact must have a lowercase `sha256` digest
   and detached signature; `latest` is not an exportable reference.
2. Copy only the immutable archives, the air-gap manifest, the public
   verification-key set, the signed offline license file, and this runbook to
   the transfer medium. Do not export PostgreSQL customer data, repository
   contents, Kubernetes Secret bytes, kubeconfigs, tokens, or private keys.
3. Record the source registry paths and a private-registry remap in the
   manifest. Remapping changes references only; it does not change digests.

## Import and verify before install

1. Disable outbound network access for the import job and installation
   namespace. The network-denied verification gate must pass.
2. Verify the manifest signature with a trusted public key, then verify every
   archive digest and detached artifact signature. Do not invoke Helm, a
   registry client, or a license provider before verification succeeds.
3. Apply the registry remap to the verified manifest and ensure every resulting
   path is an internal registry path without a URL scheme or `latest` tag.
4. Install the signed offline license from the local file using the tenant
   binding. The signing private key is never present on the target.
5. Install the exact charts/images from the verified local registry and run
   the post-install health and ownership checks.

## Rollback and evidence

Keep the previous verified manifest and registry set until the new installation
passes its health gate. Rollback selects that immutable set; it does not delete
customer environments or rewrite source credentials. Store only manifest IDs,
digests, key IDs, verification results, and timestamps in the audit record.
