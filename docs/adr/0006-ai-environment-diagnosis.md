# ADR 0006: Read-only environment diagnosis

The diagnosis API is an asynchronous, tenant/project-authorized read-only
operation. It builds context through the versioned AI context builder and
stores only bounded, schema-validated diagnosis metadata in `AIRun`; raw
context, prompts, provider responses, credentials, and commands are excluded.

`POST /api/v1/environments/{id}/diagnosis` is idempotent by key. A new run is
created only with explicit `refresh: true`. `GET
/api/v1/environments/{id}/diagnosis/{runID}` requires the same tenant, project,
and environment binding. Provider failures, disabled configuration, and
insufficient evidence return a deterministic degraded result and do not alter
environment, cluster, job, or SCM state.
