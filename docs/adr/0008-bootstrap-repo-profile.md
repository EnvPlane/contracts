# ADR-0008: Safe bootstrap repository profile

## Status

Accepted for EP-AI-013.

## Decision

Bootstrap scan produces a typed structural profile from an authorized SCM tree
projection. The allowlist includes manifest path/type, declared ports,
environment variable names, tenant component catalog IDs, and bootstrap field
names. File values and free-form text are excluded before the profile is
constructed. README content and branch names are never evidence fields.

The purpose is `bootstrap.scan`, is checked after project authorization, and is
disabled by the default disabled AI policy. It does not invoke a model, mutate
the bootstrap session, or grant additional SCM permissions.

## Non-goals

- Raw file retrieval, arbitrary manifest parsing, or secret discovery.
- Credentials, tokens, environment values, README/comments, or branch text.
- Automatic wizard progression or configuration proposal generation.
