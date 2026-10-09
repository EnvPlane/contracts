# Contracts SDK uses an outdated CI toolchain

Run 37916481579 for sub-cent contract commit f76cc92 failed govulncheck: nine
reachable standard-library advisories under Go 1.25.13, fixed in Go 1.26.9.
Other direct Go consumers already declare 1.26.9. Domain/API tests and local
checks with the patched toolchain passed; this is not a precision regression.

Implementation prompt: align contracts go.mod minimum to 1.26.9, retain the
unchanged vulnerability gate, run tests/vet/SDK-generation checks and govulncheck
with GOWORK=off, then require the new main CI to pass. Publish a new immutable
contracts version rather than replacing the existing v0.1.108 tag. Pin API and
frontend to the corrected version/source commit; verify independent consumers.
