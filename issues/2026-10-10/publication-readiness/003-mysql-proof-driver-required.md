# MySQL restore must not dispatch without lifecycle proof authority

Status: open integration prerequisite, owned by main/runner; publication held.

Huygens reports that bounded SQL transport exists but concrete lifecycle
`mysqlcopy.ProofDriver` is required before integration acceptance. Neither a
heartbeat contract-version field nor a structurally valid metadata plan proves
this authority exists. This coordinator owns contracts, not Runner code.

Implementation prompt for main/Runner: implement and test the driver required by
`runner/internal/mysqlcopy` before advertising an enabled executor or dispatching
restore_mysql. Reverify approved source PVC/StatefulSet/Service/app/admin/CA Secret
UIDs, actual source Pod image ID and TLS VERIFY_IDENTITY. Provision only exact
owned target PVC/helper resources and reuse the generated Secret without changing
it. Enforce held backup lock or scoped trusted quiescence at every stage, lifecycle
CreatedAt and plan digest, bounded stream/proof checks, clean shutdown, immutable
receipt commit and UID-only cleanup. No source data/PVC mounts in helper, grants,
credentials/rows/dumps in contract/logs, or standalone Ready evidence.

Acceptance: refusal without driver/authority; stale lifecycle/UID/image/Secret/CA;
expired lock/quiescence, partial transfer, failed shutdown or cleanup, replaced
target; successful verified current receipt releases only the MySQL prerequisite.
