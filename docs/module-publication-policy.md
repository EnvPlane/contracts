# Immutable Go module publication policy

The publisher accepts only `push` and `workflow_dispatch` on
`EnvPlane/contracts` `refs/heads/main`. It checks out the event's exact SHA.
Pull requests, fork repositories, tags and other branches cannot publish.

Automatic runs compare the candidate tree against the highest canonical stable
remote release tag, not the previous push commit. A semantic change left
unpublished by a failed run is therefore included in a subsequent docs push.
Without a released baseline, even an initial docs-only tree produces an initial
release. A released candidate or an unchanged tree does not allocate a tag.

Only regular, non-executable Markdown under `docs/` and `issues/`, and the root
README, CONTRIBUTING, SECURITY, TRADEMARKS and CHANGELOG Markdown files are
documentation exemptions. The exact paths
`.github/workflows/publish-module.yaml`, `scripts/module_publication.py` and
`scripts/tests/test_module_publication.py` are publication-metadata exemptions.
These exemptions prevent the publisher fix itself from generating an unnecessary
module release. They do not exempt other scripts/workflows, licenses, data,
schemas or arbitrary new paths. Any module Go code/test, dependency, SDK,
OpenAPI/schema, mixed or unknown change releases. Symlinks, executable docs and
renames out of the exemption set release. If Go embedding is present in either
tree, classification conservatively releases rather than guessing whether
documentation is embedded module input.

Manual main dispatch explicitly publishes a new candidate even when its changes
are only documentation or publisher metadata. Dispatching the already-published
commit is idempotent; publishing the same immutable source twice is unnecessary.

Classification tests always run, including on skipped automatic publications.
Release candidates must pass `GOWORK=off` race tests and module build. Before
tagging, the serialized publisher refreshes remote main/tags, refuses a divergent
latest release and skips superseded main candidates. Stable numeric versions
allocate one patch above the latest stable tag; prereleases/malformed versions
do not influence this allocation. Tags are annotated and pushed without force.
Competing tag creation fails instead of replacing the remote tag. An ambiguous
network/Git failure is an error, not proof that a tag is absent. No existing tag
is moved or deleted.

Broader Go CI and image publication remain separate workflows: a tag alone is
not evidence that all hosted checks or runtime acceptance passed. Consumers
should bind a reviewed immutable version/source/checksum/schema manifest rather
than treat mutable latest-tag discovery as release acceptance. After an explicit
release, verify the actual remote peeled SHA, Go module sums and hosted results.
Never speculate about the next version or update consumers before it exists.

Regression command (local temporary Git repositories; no public tag writes):

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover \
  -s scripts/tests -p 'test_module_publication.py' -v
```
