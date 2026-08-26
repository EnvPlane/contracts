# Secret materialization constants trigger gosec G101

## Problem

Gosec interpreted Secret lifecycle error and action identifiers as hardcoded
credentials because their names contain `secret`.

## Resolution

Mark the two non-sensitive constant/test values with narrowly scoped G101
suppression comments. They are protocol labels and contain no credential data.
