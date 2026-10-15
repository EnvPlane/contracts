# Installation billing provider modes

The v1 signature format already includes provider mode and audience in every
intent and session descriptor. Test retains `envplane.install-billing.test.v1`;
the live candidate uses `envplane.install-billing.live.v1`. The Go validator,
TypeScript SDK and OpenAPI require the exact mode/audience pair. Historical
test vectors, installation IDs and frame bytes are unchanged. An otherwise
valid signature or session descriptor cannot be replayed across these modes.

Validation does not choose or authorize runtime mode. A consumer must compare
scope with its protected immutable mode/account/merchant/installation/tenant
configuration before creating a challenge, session or financial effect. No
customer body, metadata, redirect or credential prefix selects that configuration.
This additive contract does not enable native live financial writes or supply
merchant/tax/custody/provider approval.

PDB-004-F02 source preparation covers Go/TypeScript/OpenAPI and regeneration of
the SDK digest. Immutable main-only module publication and exact direct-consumer
pins must follow review; source fixtures are not a released-module readback.

Verification: `GOWORK=off go test -race ./...`, `go vet ./...`,
`scripts/check-sdk-generated.sh`, `node scripts/test-installation-mode.mts`,
`node scripts/test-installation-paid.mts`. Both original test corpora remain
unchanged. New cases cover live proofs, independent rotation signatures,
mode/audience mismatch, descriptor mismatch and cross-mode proof replay.
