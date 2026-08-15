# Contributing boundary rules

EnvPlane public repositories are the compatibility core. New code must follow
ADR-0001 and ADR-0002:

- public contracts, OpenAPI, SDKs, Agent, Runner, Webhook, GitOps, Bootstrap,
  Community control-plane and Community UI stay in public repositories;
- future Enterprise-only functionality belongs in a separately versioned
  module or service and may depend only on public contracts;
- public packages must not import private modules or fetch private code at
  runtime;
- security-sensitive cluster-side and webhook code must remain inspectable;
- test fixtures must use explicit build tags and must not enter production
  binaries.

Every new Enterprise boundary change requires an ADR update and a compatibility
entry before implementation is merged.

