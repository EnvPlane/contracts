#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
policy="$repo_root/docs/open-core-policy.json"
adr="$repo_root/docs/adr/0003-open-core-repository-policy.md"

test -f "$policy"
test -f "$adr"

jq -e '
  .schemaVersion == 1 and
  .policyId == "EP-COM-003" and
  .canonicalProductName == "EnvPlane" and
  (.publicRepositories | type == "array" and length == 9) and
  ([.publicRepositories[].name] | sort) == [
    "agent", "bootstrap", "contracts", "control-plane", "deploy",
    "frontend", "gitops", "runner", "webhook"
  ] and
  all(.publicRepositories[]; (.name | type == "string" and length > 0) and
    (.responsibility | type == "string" and length > 0)) and
  (.privateModuleBoundaries | type == "array" and length > 0) and
  all(.privateModuleBoundaries[]; (.id | type == "string" and length > 0) and
    (.dependsOn | index("contracts") != null) and
    (.dependsOn | index("extension-interfaces") != null)) and
  (.dependencyRules.privateMayDependOnPublic == true) and
  (.dependencyRules.publicMayDependOnPrivate == false) and
  (.dependencyRules.publicMayFetchPrivateDynamically == false) and
  (.dependencyRules.privateForkOfPublicCoreAllowed == false) and
  (.dependencyRules.enterpriseFailureMayBlockCommunityFlow == false) and
  (.dependencyRules.plaintextCredentialsAcrossBoundary == false) and
  (.dependencyRules.crossTenantDataAcrossBoundary == false) and
  (.extensionInterfaces | type == "array" and length >= 6 and
    all(type == "string" and length > 0)) and
  (.releaseArtifactComposition.publicCompatibilitySetFirst == true) and
  (.releaseArtifactComposition.enterpriseAdditionsSeparatelyVersioned == true) and
  (.releaseArtifactComposition.publicMetadataMayContainPrivateSource == false) and
  (.releaseArtifactComposition.credentialsInline == false) and
  (.releaseArtifactComposition.compatibilityDeclarationRequired == true) and
  (.legalCheckpoint | type == "array" and length == 6 and
    all(type == "string" and length > 0))
' "$policy" >/dev/null

for repository in contracts agent runner webhook bootstrap gitops control-plane frontend deploy; do
  rg -q -- "\`$repository\`" "$adr"
done

for required in \
  "Private Enterprise" \
  "must not import a private module" \
  "private fork" \
  "Release artifact composition" \
  "Legal checkpoint" \
  "Apache-2.0" \
  "CLA" \
  "trademark"; do
  rg -qi -- "$required" "$adr"
done

echo "open-core policy check passed"
