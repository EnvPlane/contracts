# Canonical brand prose conflicts with the current guard

Status: follow-up policy clarification; publication candidate avoids affected
product prose and uses lowercase machine repository identifiers.

The workspace BRAND_POLICY.md both prescribes canonical product capitalization
for documentation and later says CI should forbid the same capitalization.
The contracts brand guard currently rejects it anywhere on added lines, including
literal GitHub owner names in links. Initial PVC ADR deciders and the publication
baseline links triggered this gate. No machine identifiers or guard logic were
changed; brand-neutral prose and lowercase repository identifiers resolve this
candidate's failures.

Implementation prompt: reconcile the policy's canonical-brand and CI sections
with owners, then update guard classification and tests across coordinated
repositories. Distinguish approved user-facing brand prose from legacy machine
names and immutable external source references. Do not relax gates unilaterally
or rename security/upgrade-sensitive compatibility identifiers.
