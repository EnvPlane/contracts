# Contributing to EnvPlane

Thank you for helping improve EnvPlane.

## Repository boundary rules

Public contracts, OpenAPI, SDKs, Agent, Runner, Webhook, GitOps, Bootstrap, and
deployment artifacts must remain independent of private product repositories.
Public packages must not import private modules or fetch private code at
runtime. Enterprise-only functionality belongs in separately versioned private
modules or services and may depend only on public contracts. Changes to this
boundary require an ADR and a compatibility entry before implementation.

1. Open an issue before substantial design work.
2. Keep changes focused and add tests for behavior changes.
3. Run the repository's documented test and lint commands.
4. Do not commit credentials, customer data, generated secrets, or private
   infrastructure details.
5. Submit a pull request describing the problem, approach, compatibility
   impact, and validation performed.

By submitting a contribution, you confirm that you have the right to submit it
and agree that it is licensed under this repository's license. Contributions
must not introduce dependencies with incompatible licensing terms.

## Individual contributor agreement

External contributors must complete the approved individual CLA before
a pull request can merge. A maintainer verifies the off-repository signature
record and adds the `cla-signed` pull-request label; the CI gate rejects
external pull requests without it. The draft agreement is in
[`docs/legal/INDIVIDUAL_CONTRIBUTOR_LICENSE_AGREEMENT_DRAFT.md`](docs/legal/INDIVIDUAL_CONTRIBUTOR_LICENSE_AGREEMENT_DRAFT.md).
Corporate, employer-sponsored and contractor contributions are not accepted
until a separate written agreement is approved. Do not upload signed agreements
or personal data to GitHub.

Security vulnerabilities must follow [SECURITY.md](SECURITY.md), not the public
issue tracker.
