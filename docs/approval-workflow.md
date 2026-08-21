# Approval workflow

Policy-marked production-like, large, and privileged-repair actions create a
tenant-scoped approval request before execution. Requests use the states
`pending`, `approved`, `rejected`, `expired`, and `canceled`.

Approval has separation of duties: the requester cannot approve their own
request. Expiry is checked against UTC time at approval and an expired request
cannot resume a job. Approval and resume are idempotent: repeating the approve
command returns the existing approved request and does not enqueue another job.
Only the requester can cancel a pending request. Cleanup/delete actions never
require approval.
