# Docs-only pushes allocate unnecessary module releases

Status: implemented and locally verified; push held for shared verification.

## Evidence and impact

The `e69da406` ticket-only main push allocated v0.1.111 even though its
domain, SDK, OpenAPI and module files were unchanged from v0.1.110. Consumer
latest-version checks then failed and required nine dependency updates.

## Implementation prompt

Restrict changes to the contracts publisher and supporting scripts, tests and
documentation. Skip genuinely documentation-only automatic publications,
preserve explicit manual dispatch on canonical main, and publish code,
dependency, OpenAPI/schema and unknown changes. Compare against the latest
immutable release so a failed earlier semantic publication is not lost.
Exempt only the exact publisher workflow/classifier/regression-test paths as
publication metadata; unrelated scripts and workflows remain release-triggering.
Test initial repositories, mixed changes, renames, trust boundaries, stale
main runs, reruns, numeric version allocation and immutable tag collisions.
Never force or rewrite a tag. Do not change domain/OpenAPI or consumer files.

## Acceptance

- A released baseline followed only by regular documentation produces no tag.
- Initial and semantic/mixed changes publish; manual main dispatch is retained.
- Forks, PR events and non-main refs cannot publish.
- Serial allocation refreshes remote refs; stale runs and reruns do not create
  duplicate versions; a competing tag push fails without overwriting it.
- Regression tests pass locally; no push until shared verification.

## Local verification

26 Python regression tests passed using temporary local Git repositories,
including competing tag rejection and missing-remote failure. Actionlint
v1.7.12 accepted the modified workflow. Released-mode Go 1.26.9 race tests,
vet, build, SDK-generated consistency, changed-file brand guard and diff checks
passed. Domain, SDK, OpenAPI and module dependency files remain unchanged.
The classifier's exact metadata exemption permits this pipeline-only iteration
to be pushed later without automatically allocating another module release.
Hosted verification and publication remain intentionally unperformed until
main's shared-verification greenlight.
