# Safe environment failure evidence for AI diagnosis

Status: implemented locally; not published.

The environment context allowlist omitted provisioning failures before Kubernetes events exist. Export only bounded typed failureCode/failurePhase and stateUpdatedAt, never raw LastError. Recognize the server-owned pre-publication executor error prefix; unknown deployment and cleanup errors remain generic redacted classifications. Ready/nonfailed records cannot retain failure evidence. All fields remain untrusted data.

Tests verify classifications, malicious credential/instruction suppression and resolved-state omission. Existing builder tests cover tenant isolation, string/entry/byte limits and central redaction. Freshness relative to the diagnosis request is enforced by the control-plane caller; stateUpdatedAt is a record timestamp, not a claimed failure occurrence time.

Codex release prompt: publish contracts only after explicit push approval; update every direct consumer from v0.1.106 to the new published version, test independently with GOWORK=off and repeat real diagnosis on a failed pre-publication environment. Local workspace testing does not prove released module pins contain this behavior. Keep the OpenAPI unchanged because no HTTP payload type changed.
