# INSTALL-BILL-006-F10: Align required CLA context with the actual protected check

Date:2026-10-14. Priority:P1. Status: source/account policy mismatch, release blocker.
Component: contracts repository protection. No AWS/provider/runtime action.

PR14 headb3a4e6b371d0a530e1ffdeec3bbbc714623c0a2c has all eight check-runs
completed/success, including Contributor agreement workflow job named cla
(run38569297378/job115869955874). Main protection requires the literal context
Contributor agreement / cla, which no check-run reports. Ordinary gh pr merge
--match-head-commit rejects the otherwise green candidate because base policy
prohibits the merge. No protection change or administrator bypass was performed.

Read repo CLA workflow and current protection/rulesets. Preserve all existing
checks/review/strict behavior. Make the required context match the actual trusted
CLA check identity, including app binding where supported. Test an exact-head PR
with passing and failing contributor agreement outcomes. Do not fabricate a
status context or weaken the CLA decision. Repository owner reviews the exact
policy correction before applying it. If using a one-off approved owner merge,
record it explicitly and keep this permanent configuration correction open.

Evidence: GET branches/main/protection requires Contributor agreement / cla;
GET commits/b3a4e6b371d0a530e1ffdeec3bbbc714623c0a2c/check-runs returns cla success.
Next step: reviewed policy alignment or explicit owner decision on one-off merge.

Scoped repair: add a real pull_request check named exactly as existing protection.
It executes the unchanged CLA script from the trusted PR base SHA, checks the same
association/verified-label policy, passes metadata through quoted environment,
uses read-only permissions and disables persisted checkout credentials. Existing
pull_request_target check and all protection settings remain unchanged. Local CLA
positive/negative fixtures pass; exact-head hosted outcome must be read back.
