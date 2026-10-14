# Installation billing contract SDK (local candidate)

Protocol `envplane.install-billing.v1`, pure Ed25519. This directory is a private,
unpublished source artifact, not an npm release or enrollment implementation.

`installation-billing.ts` mirrors Go `installationbilling`: canonical binary frames,
strict duplicate/unknown/missing/null/case-alias JSON rejection, uint64 decimal
strings, operation/path table, DTO/body bounds and stateless proof/session/response
binding. Both use `installationbilling/testdata/v1.json` (11 operations, 41 negative
vectors). Signature verification rejects noncanonical, small-order and mixed-order
A/R points and noncanonical scalars. Go verifies signatures with crypto/ed25519,
uses filippo.io/edwards25519 v1.1.0 for point validation; TypeScript uses pinned
@noble/curves and @noble/hashes 2.4.0. No custom curve implementation or signing API.

These helpers do NOT authorize a key or consume a nonce. Runtime must strongly
match persisted challenge, current registered key/epoch, session, owner and guards,
then commit admission atomically. No private keys, seller/provider credentials or
browser-selectable customer IDs are in these DTOs.

```sh
npm ci --ignore-scripts --no-audit --no-fund
npm test
```

Node 22.18+ or a compatible TypeScript runner is needed for the native test script;
crypto dependencies support Node >=20.19. Browser/Next consumers use their ordinary
TypeScript bundler. Runtime browser signing/key custody is outside this module.

Frontend copies these exact helpers and vectors with its explicit
`ENVPLANE_CONTRACTS_SOURCE=... node scripts/sync-installation-contract.mjs` command;
provenance records content hashes and says unpublished-source-integration. Existing
published module pins and snapshots must not be relabeled as this new contract.
