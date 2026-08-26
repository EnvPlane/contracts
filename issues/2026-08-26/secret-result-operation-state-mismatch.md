# Secret result operation allowed incompatible item states

## Status

Resolved in commit `b6d7808`.

## Problem

The lifecycle contract validated item result states independently from the
operation. A cleanup result could therefore be reported as `ready`, or a
materialization result as `deleted`, making executor behavior ambiguous.

## Resolution

Materialization results are restricted to `pending`, `materializing`, `ready`,
or `failed`; cleanup results are restricted to `cleaning`, `deleted`, or
`failed`.
