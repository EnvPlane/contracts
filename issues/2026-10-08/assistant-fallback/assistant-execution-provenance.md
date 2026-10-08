# Assistant execution provenance and safe observation context

Status: implemented locally; not published.

AIAssistantOutput carries ai/deterministic mode, provider/model/run metadata, safe fallback reason and validated read-only result. AIDiagnosisResult has optional executionMode/fallbackReason. Assistant snapshots are scoped to tenant/project/subject, centrally redacted, bounded and untrusted. Canonical OpenAPI and generated SDK metadata are updated; no raw executable plan or manifest is a provider input.

Codex continuation prompt: publish only after push approval, update all direct consumers and regenerate frontend API types from the published artifact. Complete route integration only after explicit authorization of the new external data scope. Verify optional metadata preserves compatibility with existing deterministic clients.
