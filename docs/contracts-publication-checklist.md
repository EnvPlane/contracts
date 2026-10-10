# Contracts publication coordination

Status: main explicitly authorized CONTRACTS-ONLY publication after final
source-profile OpenAPI/SDK/schema/released-Go/brand checks and live tag recheck.
Consumer pushes and runtime rollout remain held for SQL live/native/shared checks
and UI adapter acceptance. Snapshot inspected 2026-10-10.

## Verified remote baseline

Remote `origin` (`github.com/envplane/contracts`) is public, default branch main.
Latest observed tag `v0.1.109` is annotated and resolves to
`05dad4c60719fe91fdbb7ec3aee58ce714eb34b9`, also the observed remote main head.
[Go CI](https://github.com/envplane/contracts/actions/runs/37916751755),
[publisher](https://github.com/envplane/contracts/actions/runs/37916751825) and
[brand guard](https://github.com/envplane/contracts/actions/runs/37916751723)
passed for that SHA. These are NOT evidence for the local unpublished changes.

At initial inspection the local clean main was `ec3c162`, two commits ahead:
`a3d0c31` PVC contracts and `ec3c162` verified executor-capability metadata.
Subsequent local MySQL/schema/checklist commits must be included in the final
freeze. Re-read HEAD, status, remote head/tags and ahead log before authorization.

## Publication mechanism and safety boundary

`.github/workflows/publish-module.yaml` runs on main pushes and manual dispatch.
Automatic runs skip regular docs-only and exact publisher-metadata-only changes
since the latest immutable release; unknown and semantic paths still release.
Manual main dispatch remains explicit, with idempotence for released commits.
See [the current classification and safety policy](module-publication-policy.md).
Release candidates pass race tests and build with GOWORK=off, then allocate the
next patch from freshly fetched tags and push a new annotated immutable tag. It runs
independently of broader Go CI (vulnerability scan, vet, lint, race, coverage and
boundaries). Therefore a semantic main push is publication-triggering: do not push
to main or dispatch the publisher before main's explicit final acceptance.

Do not invent the next version, create manual speculative tags, move existing
tags or pin consumers before a real tag resolves to the frozen SHA. The publisher
chooses the version from live tag state; record its actual output. A version tag
can exist even when broader parallel CI fails; both must be green before pinning.

## Freeze and local gates

Main accepts the frozen metadata contract for publication independently of
unfinished SQL native/CEL/UI integration. The integration/live checkboxes below
remain runtime/consumer-push gates, not evidence that those tasks are complete.
A published module must never be represented as successful live SQL execution.
Local schema/checks and exact remote CI/publisher/tag verification still gate
publication and subsequent dependency updates respectively.

- [ ] Main and Huygens accept the shared source/target/service/TLS/backup/DDL
  schema and worker adapter. No data/credential payloads or automatic DB grants.
- [ ] Concrete lifecycle ProofDriver verifies source/target UID/spec/current
  lifecycle, held DDL authority, clean shutdown, immutable receipt and UID-only
  owned cleanup. Huygens reports this as required; do NOT dispatch without it.
- [ ] Independent SQL heartbeat version/scope/helper/target-image fields and
  optional Secret OutputUID are implemented by authenticated consumers.
- [ ] Source UID scans cover PVC, StatefulSet, Service and Secret. Missing
  production classification, reviewed image, CA or backup authority fails closed.
- [ ] Main/Worker derive TLS connection host exactly as service.namespace.svc or
  service.namespace.svc.cluster.local from the bound Service/namespace. Arbitrary
  suffixes and custom cluster domains are refused before credential use.
- [ ] Actual generated target Secret UID is acknowledged by the trusted current
  Secret-materialization command/attempt before compiling the MySQL plan.
- [ ] Trusted mysql-password-v1 generator independently produces app/root
  credentials. TargetRootPasswordKey is the fixed MYSQL_ROOT_PASSWORD key,
  distinct from the app key and matched by trusted target-Secret metadata.
  Worker initialization never reuses the app password or overwrites the Secret.
- [ ] Main accepts real MySQL live proof with the final driver and independently
  generated root/app credentials, including refusal and cleanup cases. This
  coordinator performs no live actions; local contracts checks are not live proof.
- [ ] Before template revision sealing, generated feature SQL workload metadata
  binds target PVC `subPath: mysql` (MySQLRestoreTargetDataSubPath), the reviewed
  pinned MySQL image and generated app/root Secret refs. Worker target layout and
  both backend publication gates match those digest-covered bindings; original
  source root mounts remain untouched. Refuse a root mount or later binding drift.
- [ ] Main lifecycle wait/dispatch/evidence paths bind current CreatedAt and
  immutable plan digest; no prior lifecycle result can release a new workload.
- [ ] Existing PVC/fake-state guards and both backend publication prerequisites
  remain intact. MySQL/PVC verified evidence is never standalone Ready.
- [ ] Freeze exact SHA and clean own files, preserving parallel contributors.
- [ ] With `GOWORK=off` and Go 1.26.9 (from go.mod): tests, race, vet, module build,
  pinned v2.12.2 lint, govulncheck v1.7.0, coverage >=30%, generated SDK consistency,
  domain/public-module/brand/shared-file boundaries all pass. Store coverage in
  a temporary file, not the tracked cover.out. Re-run on final frozen HEAD.
- [ ] Main gives final publication greenlight for that exact frozen SHA.

## Authorized publication and consumer pins (future; not executed)

1. Recheck live remote head/tags and local ahead commits; do not overwrite remote
   advancement or publish other unreviewed parallel changes.
2. After greenlight only, push the approved main SHA through the existing normal
   workflow. Observe publisher and broader CI on that exact SHA; do not dispatch
   redundant publisher runs or manufacture a version.
3. Capture actual release tag. Verify its peeled remote commit is the frozen SHA,
   publication and all required checks succeeded, and Go module resolution with
   GOWORK=off can download that exact tag (no local replacement/workspace masking).
4. Enumerate EVERY direct Go consumer, including any supplied PRIVATE workspace.
   At the initial public snapshot control-plane, runner, agent, gitops, bootstrap,
   webhook and activation-issuer all pin v0.1.109. Deploy/test-app are not direct
   consumers. No consumer pins were changed in this preparation.
5. Update only each consumer's go.mod/go.sum to the verified published tag.
   Preserve unrelated edits; do not change frontend package.json or generated API
   files. Main owns frontend generation and code changes. Do NOT push consumer
   repositories until main accepts SQL live/native/shared checks and UI alignment.
6. Use GOWORK=off for consumer validation, check that replacements do not mask
   publication, verify checksum resolution, and run each consumer's relevant
   tests/build. If typed integration fails, hand code failures to main/worker;
   coordinator owns dependency files only. Commit own pin changes in English.
7. Run `ENVPLANE_CONTRACTS_VERSION="$RELEASE_TAG" bash
   contracts/scripts/check-contract-version.sh "$WORKSPACE_ROOT"` with the ACTUAL
   verified tag and explicit root. Check PRIVATE separately if outside that root.
8. Report exact tag/SHA, hosted gates and consumer local/remote status distinctly.
   No live rollout is authorized by module publication.

## PRIVATE workspace and rollback

PRIVATE workspace was not found at the provided/expected project, document or
Dropbox locations or in saved project metadata. Main must provide its exact
root before complete private-consumer enumeration can be claimed. The nested
PRIVATE fixture in guard tests is synthetic evidence only, not a real inventory.

Before consumer mutations record original pin/checksums and own diffs. If a new
release fails acceptance, stop rollout and retain/revert only own consumer pin
changes with a normal corrective commit. Never rewrite/delete the published tag
or reset other contributors' edits. A contract defect requires a new reviewed
immutable release, not reuse of an existing version.
