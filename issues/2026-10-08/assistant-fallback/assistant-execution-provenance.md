# Assistant execution provenance and safe observation context

Status: implemented locally; not published.

AIAssistantOutput carries ai/deterministic mode, provider/model/run metadata, safe fallback reason and validated read-only result. AIDiagnosisResult has optional executionMode/fallbackReason. Assistant snapshots are scoped to tenant/project/subject, centrally redacted, bounded and untrusted. Canonical OpenAPI and generated SDK metadata are updated; no raw executable plan or manifest is a provider input.

The external data scope was approved on 2026-10-08. Explicit tenant snapshots/runs are supported only for release.engineering and approved_actions, with empty project identity and no fake binding. Optional assistant fields are present in all canonical response DTOs; strict action round-trips accept them while executable hashes ignore narration. Contracts tests and lint passed.

Codex continuation prompt: publish only after push approval, update all direct consumers and regenerate frontend API types from the published artifact. Verify optional metadata preserves compatibility with existing deterministic clients and apply control-plane migration 057 during rollout.
