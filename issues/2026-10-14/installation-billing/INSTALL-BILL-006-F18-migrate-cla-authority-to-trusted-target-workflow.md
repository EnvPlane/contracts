# INSTALL-BILL-006-F18: Move required CLA authority to the trusted target workflow

Date:2026-10-14. Priority:P0 before accepting untrusted contributor PRs.
Status: source-confirmed release-policy follow-up; no current publication spoof.
Component: contracts GitHub workflow authority. Related:006-F10.

F10 restored the stale required context Contributor agreement / cla with a real
pull_request check. Actual v0.1.118 candidate71229ec ran the existing CLA script
from trusted base SHA, passed the real policy and all CI, and merged normally.
Released source a1b9c2482576c67dd8292d549f56b553933f5cb4 and its immutable tag remain
valid; no check status was fabricated and no protection was bypassed.

However, the bridge workflow declaration itself is loaded from future PR code.
A future untrusted PR can propose replacing its job with a success-only command.
Executing a trusted-base script in today's declaration does not make future
candidate-controlled declarations an independent CLA authority. Current inspected
main protection requires the exact context, no required PR reviews and no app
binding. No malicious PR was submitted or accepted to demonstrate this condition.

Codex-ready implementation: read AGENTS/README and current branch protection,
CLA target/bridge workflows and release classifier. Preserve existing protection,
read-only permissions and verified label/association policy. In a reviewed first
phase assign the exact protected job name to the existing pull_request_target
CLA job, which loads policy from trusted base. Retain a genuine bridge only until
that target job name is present on main. In a second reviewed phase remove the
candidate-controlled pull_request bridge. Negative proof must show a PR editing
or deleting its own workflow and lacking verified CLA cannot satisfy the trusted
target check; approved association/verified-label outcomes remain truthful.
Do not fabricate statuses, disable required checks, use administrator bypass or
run candidate scripts with trusted credentials. Record exact workflow/base/head,
negative and positive outcomes and actor/label scope without credential values.

Both phases require explicit reviewed release scope: this repository's conservative
main publisher may create additional immutable module versions for workflow changes.
Do not trigger unrelated v0.1.119/v0.1.120 publications or rewrite v0.1.118 from this
ticket alone. Recheck current tags/policy before choosing the future sequence.
No AWS/provider/installed data/secret/deployment action is authorized or needed.

Safe state: v0.1.118 publication evidence is genuine; do not certify future foreign
PR CLA authority from the temporary bridge. Next step: owner reviews the two-phase
trusted-target migration and its release scope before additional main merges.
