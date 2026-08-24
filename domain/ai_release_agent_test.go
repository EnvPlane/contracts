package domain

import (
	"strings"
	"testing"
	"time"
)

func TestAIReleaseObservationRejectsMutableTag(t *testing.T) {
	o := AIReleaseObservation{TenantID: "tenant-a", UmbrellaVersion: "1.0.0", Artifacts: []AIReleaseArtifact{{Name: "api", Repository: "registry/api:latest", Digest: "sha256:" + strings.Repeat("a", 64), Present: true}}}
	if err := o.Validate(); err == nil {
		t.Fatal("mutable tag must be rejected")
	}
}
func TestAIReleasePlanRequiresApprovalForPublication(t *testing.T) {
	now := time.Now().UTC()
	p := AIReleaseAgentPlan{SchemaVersion: AIReleaseAgentSchemaVersion, PlanID: "p", TenantID: "tenant-a", ContextHash: "h", GeneratedAt: now, Observation: AIReleaseObservation{TenantID: "tenant-a", UmbrellaVersion: "1", Artifacts: []AIReleaseArtifact{{Name: "api", Repository: "registry/api", Digest: "sha256:" + strings.Repeat("a", 64)}}}, Evidence: []AIEvidenceReference{{TenantID: "tenant-a", SourceID: "release:1"}}, Proposals: []AIReleaseRepairProposal{{Kind: AIReleasePublishUmbrella, Repository: "org/repo", Branch: "main", Tests: []string{"contract"}, DigestRequired: true}}}
	if err := p.Validate(); err == nil {
		t.Fatal("publication without approval must fail")
	}
}
