# MySQL TLS host suffix could bypass Service identity

Status: fixed locally for initial cluster.local support; custom-domain support
is deferred and must not bypass this publication gate.

Security review found that prefix validation accepted service.namespace.svc plus
arbitrary DNS suffixes. Worker uses TLS ServerName as the connection host, so an
attacker endpoint could receive approved credentials despite Service UID proof.

Current fix: require exact service.namespace.svc or
service.namespace.svc.cluster.local in CanonicalIdentity; schema descriptions,
pattern and tests enforce this boundary even when plan/trusted source agree.
Main UI/source policy and Worker must enforce the same exact derivation. No
publication until this fix and integrated refusal tests are accepted.

Future implementation prompt: if custom cluster DNS is needed, design a trusted
cluster-domain binding derived from authenticated cluster identity, with provenance,
exact Service DNS derivation and contract/digest/authority checks. Never accept a
caller-supplied suffix, profile flag, arbitrary TLS hostname or insecure mode.
Test credential refusal for attacker suffixes, changed namespace/Service, invalid
domain provenance and replaced Service UID. Preserve existing fail-closed v1.
