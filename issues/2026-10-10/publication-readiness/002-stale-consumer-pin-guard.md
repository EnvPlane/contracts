# Consumer version guard uses a stale default and omits direct consumers

Status: fixed locally; release use still requires a verified published tag.

Evidence: default `v0.1.88` rejects all six original public consumer pins at
`v0.1.109`, although that tag exists on origin. Hardcoded enumeration omits
activation-issuer and any nested PRIVATE workspace consumers. Standalone remote
contracts CI has no assembled siblings and therefore did not expose this drift.

Implementation prompt: require explicit ENVPLANE_CONTRACTS_VERSION when direct
consumers exist; never infer or guess a future release. Discover direct contracts
dependencies from all go.mod files under an explicitly provided workspace root,
including nested/private modules, excluding vendor/node_modules/.git. Keep
standalone contracts-only CI valid. Add tests and verify the seven existing
public consumers using the confirmed remote v0.1.109 tag without changing pins.
