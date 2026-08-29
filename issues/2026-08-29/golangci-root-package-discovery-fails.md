# golangci-lint root package discovery fails before analysis

## Problem

`scripts/golangci-lint.sh` invokes `golangci-lint run ./...`, but the selected
v2 binary exits with `no go files to analyze` even though `go test ./...`
discovers the root package and all subpackages. This prevents CI lint from
reaching source diagnostics.

## Expected

The lint wrapper must analyse the module packages using the same package set as
`go test ./...`, without requiring a root `main` package.

## Impact

EP-SSO-007 contract changes were validated with targeted domain tests; the
wrapper failure remains independently tracked.
