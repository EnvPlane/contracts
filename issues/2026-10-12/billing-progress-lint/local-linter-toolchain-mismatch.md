# Linter cache compatibility with newer effective Go

Status: local verification unblocked by explicit CI toolchain; script hardening pending.
Priority: P2.

## Problem
The unpinned local lint command panicked: file requires Go1.27 while linter was
built with Go1.26. scripts/golangci-lint.sh currently validates builder against
module minimum, not necessarily the effective Go analyzing standard packages.
Pinning GOTOOLCHAIN=go1.26.9 reproduced CI and passed with zero issues. This is
separate from the hosted QF1001 source failure, which was fixed minimally.

## Codex implementation prompt
Read AGENTS and scripts/golangci-lint.sh. Validate cached linter builder against
both module minimum and effective selected Go version. Rebuild only when needed;
retain version pin and automatic toolchain resolution for older system Go.
Add shell-fixture coverage for newer selected Go versus cached older builder and
compatible reuse; run actual lint. Do not suppress diagnostics or modify billing
contracts. English commit; existing push authorization applies during this task.
