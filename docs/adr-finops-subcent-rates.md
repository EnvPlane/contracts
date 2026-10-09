# ADR: exact sub-cent resource rates

Status: Accepted for implementation
Date: 2026-10-09
Deciders: project maintainer request; implementation by Codex

## Context
Whole minor-unit rates cannot express EUR 0.005/GiB-hour RAM or EUR
0.00014/GiB-hour storage. Existing money, budgets and old price snapshots must
retain integer semantics. Missing prices must never become known zero prices.

## Decision
Use additive scaled rate objects: minorUnits / 10^scale, scale 0..6 and an
effective maximum of 1,000,000,000 minor units per resource unit. Preserve
integer legacy fields and forbid simultaneous legacy/scaled representations.
CPU/RAM legacy fields are null for scaled rates. A scaled dimension has
legacy known=false and minorUnitsPerUnit=0, so older readers fail closed as
UNKNOWN rather than interpreting a numerator or shadow zero as a real price.
New readers derive price knowledge from a validated scaledRate object.

Use rational multiplication and cumulative half-up rounding to integer money;
never independently round each short measurement or retroactively reprice old
versions. Budgets and infrastructure-report amounts remain integer minor units.
Reject unsafe aggregate money instead of exposing unsafe JavaScript integers.

## Options considered
Binary floating-point rates are simple but unsuitable for exact monetary
arithmetic. Decimal strings require extra parsing and canonicalization in every
consumer. A bounded integer/scale object is explicit, exact and portable.

## Consequences and actions
Update domain/OpenAPI, API validation/persistence, both allocation paths and UI
atomically. Old integral snapshots stay readable and unmodified. Strict old APIs
reject new fields; downgrades cannot price scaled snapshots accurately and must
be treated as UNKNOWN. No DB payload rewrite, budget creation, provider invoice
claim or requested/provisioned storage double counting is introduced.
