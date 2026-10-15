# PAY-005-F03 Checkout URL fragment compatibility

This source candidate stacks on the mode/audience contract change `5d98d08`.
It is neither a published module nor actual Stripe readback or delivery evidence.
Issuer native validators and immutable consumer pins require a separately reviewed
integration after the compatible contract is published.

Only a ready `checkout` response may carry an opaque fragment on the exact
`https://checkout.stripe.com/` origin. Go and TypeScript preserve the entire
provider URL byte-for-byte, including escaped fragment bytes, without decoding
it or changing request/proof/session binding. HTTPS, exact origin, no userinfo
or explicit port, 4096 UTF-8 byte limit, and no raw whitespace/control/backslash
remain admission requirements. Malformed fragment percent escapes are rejected
in both implementations. A fragment grants no payment or license authority.

Portal responses retain the fragment prohibition, including an empty `#`.
Checkout/portal requests still accept only the approved `billing` return route;
caller URL fields and fragments in that route remain invalid. No server return
URL policy is broadened by this response-only compatibility change.

`installationbilling/testdata/checkout-url-v1.json` contains 25 synthetic shared
vectors, including opaque fragment preservation, forbidden portal fragments,
foreign/return origins, userinfo, ports, control bytes and malformed escapes.
`TestCheckoutOpaqueURLFragments` and the TypeScript contract suite additionally
exercise exact length boundaries, request correlation and return route rejection.
The historical request/signature corpus is unchanged, including its mode/audience
and session isolation checks. These fixtures contain no real session or token.

Never log a hosted URL or fragment to diagnose rejection; report a bounded error
code and nonsecret attempt correlation instead. This change makes no provider
request and does not prove any current provider URL shape, deployed consumer
behavior, browser navigation, or production readiness. PDB-003/GF-013 owns
separately authorized actual provider acceptance.
