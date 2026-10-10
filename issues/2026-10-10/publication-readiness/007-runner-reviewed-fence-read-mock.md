# Released consumer test double rejects cluster-scoped admission reads

Status: RESOLVED local regression. Product pushes and final shared live acceptance
remain held. Contracts v0.1.110 publication and hosted Go CI passed at f015966.

Initial released-module validation with GOWORK=off and Go 1.26.9 built Runner, but
`go test -count=1 ./...` failed TestReviewedAdmissionFencePinsUIDSpecAndSource
in exact, policy-uid, spec-sha, source-name and stale-generation cases.
The test double reports "non-read metadata operation" for the cluster-scoped
GET of validatingadmissionpolicies.admissionregistration.k8s.io by exact name.
No coordinator changes were made to Runner code; only go.mod/go.sum are in scope.

## Resolution and evidence, 2026-10-10

Aquinas aligned the reviewed-fence test with the existing `fenceTestCommands`
fake, which supports exact cluster-scoped admission metadata reads. The test
continues to reject mismatched policy UID/spec/source and stale generation;
the exact case still requires effective source probes.

- Main reports fresh `GOWORK=off go test -race ./...` PASS for full Runner and API,
  including the reviewed fence cases, and lint with zero issues. These full-run
  results are main-reported evidence, not a new coordinator full-suite run.
- Coordinator independently reproduced `GOWORK=off GOTOOLCHAIN=go1.26.9
  go test -race -count=1 ./internal/pvccopy
  -run '^TestReviewedAdmissionFencePinsUIDSpecAndSource$'`: PASS, 1.175s.
- Runner HEAD at this check: e969d87198e81d7e63f413d7330aad3f119751b5 (released
  contracts dependency pin). The reviewed-fence test/implementation files remain
  untracked working-tree changes; no committed fake-fix SHA is asserted here.
- API HEAD observed: a36481c72c9d6a9602240a2776c8f161e5527977. All nine discovered
  consumer manifests still pass the v0.1.110 pin guard without a contracts replace.

This resolves the local regression only. It is not evidence that shared/native
live acceptance is complete and does not authorize any consumer/product push.

Completed implementation prompt for Runner owner: align the read-only test command double
with the reviewed fence verifier's exact cluster-scoped admission-policy read.
Do not allow mutation, arbitrary resource/name reads or weaken UID/spec/source/
generation assertions. Distinguish legitimate cluster-scoped GET from unsafe
operations. Run the focused fence cases, then full released-module builds/tests
and shared/native/live gates before authorizing product pushes.
