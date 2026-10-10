# Released consumer test double rejects cluster-scoped admission reads

Status: consumer code-owner follow-up; product pushes remain held. Contracts
v0.1.110 publication and hosted Go CI passed at f015966.

Released-module validation with GOWORK=off and Go 1.26.9 builds Runner, but
`go test -count=1 ./...` fails TestReviewedAdmissionFencePinsUIDSpecAndSource
in exact, policy-uid, spec-sha, source-name and stale-generation cases.
The test double reports "non-read metadata operation" for the cluster-scoped
GET of validatingadmissionpolicies.admissionregistration.k8s.io by exact name.
No coordinator changes were made to Runner code; only go.mod/go.sum are in scope.

Implementation prompt for Runner owner: align the read-only test command double
with the reviewed fence verifier's exact cluster-scoped admission-policy read.
Do not allow mutation, arbitrary resource/name reads or weaken UID/spec/source/
generation assertions. Distinguish legitimate cluster-scoped GET from unsafe
operations. Run the focused fence cases, then full released-module builds/tests
and shared/native/live gates before authorizing product pushes.
