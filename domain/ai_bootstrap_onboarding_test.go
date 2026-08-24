package domain

import (
	"testing"
	"time"
)

func TestAIBootstrapOnboardingRequiresEvidenceAndManualAuthority(t *testing.T) {
	plan := AIBootstrapOnboardingPlan{SchemaVersion: AIBootstrapOnboardingSchemaVersion, PlanID: "plan-1", TenantID: "tenant-a", ProjectID: "project-a", BootstrapSessionID: "session-a", ContextHash: "ctx", ManualAuthoritative: true, ReadyForReview: true, GeneratedAt: time.Unix(1, 0).UTC(), Discovery: AIBootstrapDiscovery{RepositoryProvider: "github", Repository: "org/app"}, Suggestions: []AIBootstrapFieldSuggestion{{Field: "repository.default_branch", Value: "main", Evidence: []AIEvidenceReference{{SourceType: "repo_profile", SourceID: "repo-profile:project-a", TenantID: "tenant-a"}}, Confidence: AIConfidence{Score: .8, Level: AIConfidenceMedium, EvidenceCount: 1}, RequiresReview: true, Untrusted: true}}}
	if err := plan.Validate(); err != nil {
		t.Fatal(err)
	}
	plan.Suggestions[0].Evidence[0].TenantID = "tenant-b"
	if err := plan.Validate(); err == nil {
		t.Fatal("cross-tenant evidence must be rejected")
	}
}
