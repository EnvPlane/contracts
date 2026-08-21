# ADR-0003: HA topology and disaster recovery boundary

## Decision

The control plane runs as stateless API replicas behind a health-checked
service. PostgreSQL is the authoritative state store and uses its managed HA
and point-in-time recovery facilities. Redis/queue workers are disposable and
recreated from durable PostgreSQL state. Kubernetes metadata required for
ownership and reconciliation is backed up as a separate immutable object.

Backup manifests are tenant-scoped and contain object references, project IDs,
and required subscription/audit state references. They never contain external
Secret bytes, provider credentials, kubeconfigs, signing private keys, or
license private material. A restore into an empty target restores durable
tenant/project/subscription/audit state first, then requires explicit rebinding
of each external Secret reference before runtime reconciliation is enabled.

## Restore and verification

Restore is declarative and must be verified before production cutover. The
verification report records measured restore duration as RTO and the time from
the source observation to verification as RPO. Tenant, project, subscription,
and audit consistency checks are mandatory. A drill schedule is expressed in
hours and UTC timestamps; it is metadata only and does not silently mutate
production resources.

## Failure and rollback

An invalid manifest, cross-tenant restore request, missing consistency check,
or unresolved Secret binding blocks the restore. Rollback returns to the
previous database/PITR target and immutable Kubernetes metadata references;
external Secret bindings are re-established rather than copied. No restore or
DR drill deletes source environments, source PVCs, source databases, or source
credentials.
