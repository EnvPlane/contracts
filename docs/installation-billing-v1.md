# Installation billing v1 contracts and release sequence

Date: 2026-10-14. Status: locally implemented contracts candidate, unpublished.
Base contracts source: 496c3c1628b86f610c760b465a267d2b628294a1 (v0.1.117).
Wire version: envplane.install-billing.v1. OpenAPI document version stays 1.0.0
(additive paths/schemas). No new module tag selected or published.

Go source: installationbilling; TypeScript source: sdk/typescript. Existing
ActivationGrant/LicenseGrant serialization and signatures are unchanged. Public
fixtures use deterministic byte-range test seeds only. No actual secret/provider
or installed data access occurred.

## Closed DTOs and exact routes

All operations use POST under `/v1/installation-billing/`; content encoding is
identity, complete envelope <=65536 bytes, decoded body <=32768 bytes. Challenge
uses `/challenge`. Operation suffixes:

| operation | path suffix | decoded body |
|---|---|---|
| enroll | /enroll | {} |
| session.checkout / session.portal / session.status | /sessions/checkout / /sessions/portal / /sessions/status | {} |
| checkout | /checkout | attemptId, planId, planVersion, returnRoute="billing" |
| portal | /portal | returnRoute="billing" |
| status | /status | attemptId, cursor, limit (integer 1..25) |
| tenant.bind | /tenants/bind | {} |
| key.rotate | /keys/rotate | {} (keys/epoch are signed intent fields) |
| key.revoke | /keys/revoke | {} |
| key.status | /keys/status | transitionRequestId (canonical base64url 16 bytes) |

No other operation is admitted; public redemption/commands remain future scope.
Routes are contracts only; they are not registered in issuer/control-plane or IaC.
ChallengeRequest and Request use explicit version/algorithm/intent and base64url
body. Request additionally carries durable challenge, signature, nextSignature and
sessionHandle; the last two are empty except rotation and session actions,
respectively. All fields are required; no arbitrary customerId, token, endpoint,
reference, provider flags or return URL. 32-byte key/digest fields use canonical
unpadded base64url. Genesis ID is full SHA-256 framed key, prefix epi1_ed25519_.
Scope fixes seller/merchant/test/audience/installation/tenant. Actual deployment
merchant must match approved configuration, not the synthetic acct_fixture.

Numeric epochs/timestamps are decimal uint64 strings, epoch <=2^63-1. Binary frame
fields remain u64be. No float/exponent/leading-zero aliases. Clock/TTL: exactly 300s,
30s issuing-clock tolerance, no expiration grace. Full intentDigest is the durable
request fingerprint, requestId unique across operation/tenant within installation.
Exact body bytes are hashed; no JSON reserialization. A client preserves serialized
bytes across retry. Responses select exact operation/request/digest/key epoch;
MatchResponse rejects foreign/delayed results. Response states are closed and do
not assert payment from a redirect. Unused session/url/records fields are explicit.

Pure Ed25519 over proof frames only. Both verifiers require canonical non-small-
order torsion-free public and R points, canonical scalar, exact lengths. This
strict profile intentionally does not accept ZIP-215 aliases. Point validation
uses reviewed libraries; curve arithmetic is not implemented in this project.
Stateless ParseVerifiedRequest still is NOT durable authorization.

## Errors and transport

Errors contain only version, allowlisted code and requestId (empty if unavailable).
400: invalid_request, unsupported_version, unsupported_algorithm, invalid_proof.
403: scope_rejected, authority_revoked, key_epoch_stale, session_rejected.
404: not_found (only after scoped authorization). 409: request_conflict,
challenge_consumed. 410: challenge_expired, session_expired. 429: rate_limited.
503: temporarily_unavailable. Unknown/live/foreign admission must not disclose
whether an unrelated customer/tenant exists. Error mapping is bounded, no raw
provider message/body. Checkout 202 means accepted durable work; fresh status is
needed for canonical completion. Retry-After is bounded by future handler policy.

OpenAPI marks these operations x-envplane-host=seller-issuer and implementation=
contract-only. Control-plane's existing served snapshot remains at its published
pin until coordinated publication; its route parity checker now validates the
exact 12-operation seller inventory before excluding it from CP handler coverage.
A host marker cannot hide an arbitrary control-plane route. Deployment must still
register only explicitly reviewed public routes, never the private mux.

## Local checks and future publication

Run Go race/vet/build and scripts/check-sdk-generated.sh; regenerate/check fixtures
with `go run ./cmd/installationvectors -check`; npm ci/test in sdk/typescript; compile
helpers under frontend target ES2017 with strict TypeScript; run frontend generation,
source hash checks, typecheck, unit/lint. Issuer and control-plane use build-tagged
installation_contract tests against a disposable go.work explicitly pointing to
this source. This is local source integration, not published-consumer acceptance.
No permanent replace or go.work is committed.

Publication is explicitly deferred by user instruction. After separate command:
1. Recheck tags/release workflow; select a new compatible version, never move a tag.
2. Publish only reviewed contracts source, verify module sums and exact source SHA.
3. Update activation-issuer/control-plane pins and contracts-release.json; synchronize
   CP served artifact via its guarded release workflow, not by falsifying snapshot.
4. Update frontend openapiContractRef and generate API/helper/vector artifacts from
   that exact release; verify provenance and lockfile. Current ref remains v0.1.117;
   source candidate artifacts are separately marked unpublished in their manifest.
5. Discover every direct Go consumer, including agent, bootstrap, gitops, webhook,
   runner, deploy and nested deploy live drivers; update compatible pins/snapshots
   according to existing coordinated release policy, run GOWORK=off test/build/vet.
6. Run exact-SHA hosted gates for contracts and affected consumers; new server
   routes remain disabled until INSTALL-BILL-004 through 011 persisted/parity gates.

No AWS/provider/IAM/rotation delivery proof is claimed by this contract suite.
