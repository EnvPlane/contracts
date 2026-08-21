# EnvPlane public SDK

The Go SDK lives at `github.com/envplane/contracts/sdk/go/envplanesdk` and is
versioned independently using SemVer. Patch releases preserve wire behavior;
minor releases may add optional fields/features; breaking OpenAPI or interface
changes require a new major version. The SDK sends credentials only from a
runtime `TokenProvider`, never stores them, and supplies `Idempotency-Key` for
retries.

Pagination uses `limit` and opaque `cursor`; webhook verification is HMAC
SHA-256; capability negotiation reads `X-EnvPlane-API-Version` and
`X-EnvPlane-Capabilities`. A client compiled against v1 continues to work when
the server adds compatible minor capabilities.

## Compatibility and generation

| SDK major | API contract major | Supported server contract |
|---|---:|---|
| `v1.x` | `1` | `1.x`, including additive fields and operations |

The published SDK exposes the transport primitives required by integrations:
runtime `TokenProvider` authentication, opaque cursor pagination, typed API
errors, caller-supplied idempotency keys, HMAC webhook verification and
bounded capability negotiation. It does not expose or depend on Enterprise
implementations.

The compatibility matrix is exercised by the SDK tests for: authenticated
paginated requests, idempotent JSON mutations, additive minor responses,
structured quota errors, webhook signature verification and capability headers.

`openapi/openapi.json` is the canonical source. `go generate
./sdk/go/envplanesdk` deterministically publishes its version, SHA-256 and
method/path inventory into the SDK. CI rejects a stale generated file. A
breaking schema or operation change requires a new API contract major and SDK
major; additive changes preserve the existing major.
