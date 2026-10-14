# CONTRACTS-CLA-001: Verified Dependabot updates fail the individual CLA gate

Date: 2026-10-14
Status: Fixed locally; hosted verification pending

Evidence: contracts PR #15, author dependabot[bot], GitHub user ID 49699333,
account type Bot, association CONTRIBUTOR, labels dependencies/go. The gate
only used association and labels and therefore required an individual CLA.

Codex implementation prompt: pass PR author login, numeric ID and account type
from GitHub event metadata into both CLA workflows. Exempt only the exact
verified Dependabot identity. Preserve trusted association and cla-signed
behavior; never exempt labels, arbitrary bots or branch names. Cover missing
and mismatched identity fields with negative tests. Commit and push after local
checks; verify hosted checks for PR #15 without bypassing branch protection.

Scope: do not change the separate F18 workflow-authority migration.
