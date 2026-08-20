# 2026-08-20 open-core legal checkpoint

## Recorded decision

The workspace owner selected the following repository boundary:

- Apache-2.0 public core: `contracts`, `agent`, `runner`, `webhook`,
  `bootstrap`, `gitops`, and `deploy`.
- Commercial repositories: `control-plane` under Business Source License 1.1
  and `frontend` under its proprietary license.

This checkpoint records the repository-boundary decision only. It does not
change a license, repository visibility, release distribution or public API.

## Evidence reviewed

The seven Apache core repositories contain Apache-2.0 license text, `NOTICE`
and contribution terms. `control-plane` declares Business Source License 1.1;
`frontend` declares its proprietary license. All nine use the EnvPlane GitHub
organization remotes.

## Open legal controls

CLA/DCO governance and an EnvPlane trademark policy are not yet approved or
implemented. Before a license, visibility, artifact-distribution or trademark
change, an authorized legal owner must approve the six checklist entries in
`docs/open-core-policy.json` and record the approver, date and scope here.
