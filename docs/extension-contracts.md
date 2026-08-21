# Stable extension contracts

The public `contracts/domain` module owns the versioned extension interfaces
for entitlement resolution, identity, policy evaluation, redacted audit sinks,
and FinOps allocation. Every call carries a tenant-scoped principal where
applicable and returns public typed data. Provider credentials, Secret bytes,
private keys and customer payloads are outside these interfaces.

The control-plane Community composition is complete in the Free scope. Missing
ports are filled with Community implementations, so an Enterprise module can
replace ports independently without changing public-core imports or API
contracts. Enterprise policy failures fail closed for create/mutation actions;
cleanup, delete and export use a warning fallback so recovery is not stranded.

Public CI rejects imports or dependency-graph entries for private EnvPlane
modules. Enterprise composition is therefore a separately built consumer of
these contracts, not a hidden runtime dependency of the public repositories.
