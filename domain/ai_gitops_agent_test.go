package domain

import (
	"testing"
	"time"
)

func TestAIGitOpsPlanRejectsPathEscapeAndCrossTenantEvidence(t *testing.T) {
	base := AIGitOpsAgentPlan{SchemaVersion: AIGitOpsAgentSchemaVersion, PlanID: "plan-1", TenantID: "tenant-a", ProjectID: "project-a", ContextHash: "ctx", Observation: AIGitOpsObservation{Repository: "org/gitops", AuthorizedPath: "clusters/team-a", OutputPath: "clusters/team-a/env.yaml", Controller: "flux", ControllerStatus: "ready"}, Evidence: []AIEvidenceReference{{SourceType: "gitops", SourceID: "status-1", TenantID: "tenant-a"}}, GeneratedAt: time.Unix(1, 0).UTC()}
	if err := base.Validate(); err != nil {
		t.Fatal(err)
	}
	base.Observation.OutputPath = "clusters/other/env.yaml"
	if err := base.Validate(); err == nil {
		t.Fatal("path escape must be rejected")
	}
	base.Observation.OutputPath = "clusters/team-a/env.yaml"
	base.Evidence[0].TenantID = "tenant-b"
	if err := base.Validate(); err == nil {
		t.Fatal("cross-tenant evidence must be rejected")
	}
}
