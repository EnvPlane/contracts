#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
policy="$repo_root/docs/open-core-policy.json"
adr="$repo_root/docs/adr/0003-open-core-repository-policy.md"

test -f "$policy"
test -f "$adr"

jq -e '
  .schemaVersion == 2 and
  .policyId == "EP-COM-003" and
  .canonicalProductName == "EnvPlane" and
  (.publicApacheRepositories | type == "array" and length == 7) and
  ([.publicApacheRepositories[].name] | sort) == [
    "agent", "bootstrap", "contracts", "deploy", "gitops", "runner",
    "webhook"
  ] and
  all(.publicApacheRepositories[]; (.name | type == "string" and length > 0) and
    (.responsibility | type == "string" and length > 0)) and
  (.commercialRepositories | type == "array" and length == 2) and
  ([.commercialRepositories[].name] | sort) == ["control-plane", "frontend"] and
  (any(.commercialRepositories[]; .name == "control-plane" and .license == "BSL-1.1")) and
  (any(.commercialRepositories[]; .name == "frontend" and .license == "Proprietary")) and
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
  (.legalCheckpoint.decisionDate == "2026-08-20") and
  (.legalCheckpoint.repositoryBoundaryApproved == true) and
  (.legalCheckpoint.claPolicyApproved == false) and
  (.legalCheckpoint.trademarkPolicyApproved == false) and
  (.legalCheckpoint.requiredBeforeLicenseOrVisibilityChange | type == "array" and length == 6 and
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
