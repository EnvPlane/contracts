# ADR 0002: installation-wide Authentication Settings

## Decision

Interactive browser login is controlled by one installation-wide
`AuthenticationSettings` state machine. It is not SCM OAuth, tenant membership,
Agent/Runner/bootstrap credentials, webhook authentication, or SCIM bearer
authentication. Those identities continue to work when interactive OAuth is
disabled.

The canonical OpenAPI schemas are `AuthenticationSettings` (safe read model)
and `AuthenticationSettingsCommand` (write command). The read model physically
has no `clientSecret` field. `clientSecret` is write-only, never audited or
logged, and is stored only in the credential backend.

`mode` is `disabled` or `oauth_required`; `provider` is `github`, `gitlab`,
`oidc`, or null. `state` is `setup_required`, `setup_claimed`, or `configured`.
`configured` is true only for a validated persisted credential revision.

## Provider fields

| Provider | Required safe fields | Optional safe fields | Write-only |
|---|---|---|---|
| GitHub | `clientId` | authorization/token/user-info/revocation URL overrides, scopes | `clientSecret` |
| GitLab | `clientId` | authorization/token/user-info/revocation URL overrides, scopes | `clientSecret` |
| OIDC | `clientId`, `authorizationUrl`, `tokenUrl`, `userInfoUrl` | issuer, revocation URL, scopes (default `openid profile email`) | `clientSecret` |

The callback is derived from the canonical public control-plane URL plus
`/auth/{provider}/callback`. A command cannot supply a callback URL.

## State and authorization

| State | Enable/claim | Change provider or rotate credential | Disable |
|---|---|---|---|
| `setup_required` | first authenticated installation administrator claims setup | claimant/admin, after claim | installation administrator |
| `setup_claimed` | claimant completes or releases setup | claimant or installation administrator | installation administrator |
| `configured` | installation administrator | installation administrator | installation administrator |

First-user claim is atomic and bound to the authenticated actor; it is not a
tenant membership grant. Every mutation supplies `expectedCredentialRevision`.
A stale revision returns conflict. Validation and durable credential/settings
write are one transaction: on validation failure, Secret failure, or settings
write failure, the prior working revision remains active. Credential rotation
increments `credentialRevision`; session invalidation increments
`sessionRevision`. Disable/enable races are serialized by the same revision.

## Threat model and operations

- First-user takeover: only an authenticated installation administrator can
  claim; the claim is atomic and audited without secrets.
- Partial Secret/settings write: write staging plus compare-and-swap; rollback
  or retain the old active revision on any failure.
- Leaked credential: never return/log it; rotate credential and session
  revision, then invalidate provider-token cache.
- Session replay: signed expiring state/cookies and `sessionRevision` reject
  old sessions after disable/rotation.
- Disable/enable race: optimistic credential revision and serialized commit.
- Multi-replica cache: replicas key cache entries by credential/session
  revision and invalidate on committed revision change.

No OAuth Secret is required for the first Helm install: the installation begins
in `setup_required`/`disabled`. Legacy `auth.existingSecret` is only a
migration/bootstrap source; it is not the target runtime contract.
