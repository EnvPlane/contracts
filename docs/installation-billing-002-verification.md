# INSTALL-BILL-002 local contract verification

Date: 2026-10-14. Status: locally implemented; publication and hosted gates deferred.
Worktree /private/tmp/contracts-next-iteration-20261014; branch codex/ticket-batch-20261014.
Base 496c3c1628b86f610c760b465a267d2b628294a1 (published v0.1.117).
Wire version envplane.install-billing.v1; no new Go/npm release selected or published.

Changes: installationbilling Go package, standalone TypeScript SDK, strict JSON and
pure Ed25519 frame/proof helpers, closed body/response/error DTOs, exact 12-route
seller inventory, 26 additive OpenAPI schemas and regenerated Go SDK metadata.
Signature acceptance is canonical, non-small-order, torsion-free A/R and canonical
S, with reviewed libraries; no project-owned curve arithmetic. ParseVerifiedRequest
is stateless and cannot replace persisted challenge/key/epoch/session authority.

Fixtures: 11 positive operations, 41 negative vectors shared byte-for-byte by Go,
TypeScript and frontend. Generator checks its own negatives and freezes frames,
digests, signatures, challenge DTOs, response bindings and session descriptors.
Additional unit cases cover all byte truncation/unknown schema behavior where
applicable, body/numeric aliases, exact uint64 framing above JavaScript safe integer,
expiry/skew/overflow and delayed foreign responses. Existing activation/grant types
were not changed; full domain compatibility tests pass. Every pre-existing OpenAPI
path/schema is semantically identical to the source base; original bytes outside
new blocks retained. `git diff --minimal` shows purely additive schema/path blocks.

Local commands PASS (Go1.27.2 darwin/arm64, Node26.11.1, TypeScript5.9.3):

```sh
GOCACHE=/private/tmp/install-bill-002-gocache GOWORK=off go test -race -count=1 ./...
GOCACHE=/private/tmp/install-bill-002-gocache GOWORK=off go build ./...
GOCACHE=/private/tmp/install-bill-002-gocache GOWORK=off go vet ./...
GOCACHE=/private/tmp/install-bill-002-gocache GOWORK=off bash scripts/check-sdk-generated.sh
GOCACHE=/private/tmp/install-bill-002-gocache GOWORK=off go run ./cmd/installationvectors -check
npm --prefix sdk/typescript test
# Compile SDK with frontend TypeScript target ES2017 and strict=true.
git diff --check
```

Full Go suite required sandbox escalation for loopback httptest; rerun PASS. New
Go contract tests plus five Node tests pass with zero skips. Numeric test added
last was verified by another scoped race run. SDK test runtime needs Node22.18+;
crypto dependencies require >=20.19. No dependency lifecycle scripts executed.

Dependencies pinned/checksummed: filippo.io/edwards25519 v1.1.0 (Go module sum in
go.sum), @noble/curves and @noble/hashes 2.4.0 (npm integrity in lockfile).
Current primary API/source reviewed at pkg.go.dev/filippo.io/edwards25519 and
https://github.com/paulmillr/noble-curves; v1.1.0 point decoder behavior was also
read from pinned local module source. Synthetic seeds are public byte ranges.

Consumer source integration: separate temporary go.work selects this source for
issuer/control-plane tagged installation_contract tests. No permanent replace,
no go.mod pin updates, no rewritten published provenance. Frontend generated API,
SDK/vector copy and hash manifest explicitly say unpublished-source-integration.
Control-plane served OpenAPI and reviewed contracts-release.json remain published
v0.1.117, issuer remains v0.1.116. Main worktrees were preserved. Exact consumer
commits/artifact hashes are consolidated in issuer docs/verification/install-bill-002.md.

Actual AWS/provider/secrets/installed DB/CI/publication: NONE. Protocol handlers,
storage, rate admission and native financial guards are not implemented by these
helpers. Publication/pins/hosted checks follow docs/installation-billing-v1.md only
after separate user command; 003–011 remain runtime/release work.
