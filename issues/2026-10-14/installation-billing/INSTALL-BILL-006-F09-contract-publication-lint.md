# INSTALL-BILL-006-F09: Verify installation contracts with release lint

Date: 2026-10-14. Status: fixed locally; exact-head hosted recheck required.
Component: contracts. Found during requested INSTALL-BILL-006 publication.

PR14 Go CI run38568598409 at45a8692299f79da9324d88142cc5ecd638971592 fails
errcheck on unchecked binary.Write calls in canonical framing and unchecked
json.Unmarshal in the public synthetic vector generator; gosec rejects generated
fixture creation mode0644. No provider, deployed runtime or private key was used.

Check all generic binary.Write results and framing bounds,
check generator JSON decoding and create owned vector files with0600. Existing
canonical vectors must remain byte-for-byte unchanged. Run full module race,
build/vet/lint, SDK generation and cross-language vector checks, then exact-head
Go CI before immutable publication. Preserve failed historical run metadata.

No AWS/provider/installed database action or secret read is authorized/required.
