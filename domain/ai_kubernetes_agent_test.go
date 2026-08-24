package domain

import (
	"testing"
	"time"
)

func TestAIKubernetesObservationRejectsForeignResource(t *testing.T) {
	o := AIKubernetesObservation{TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", Namespace: "env-a", NamespaceValid: true, OwnershipValid: true, Resources: []AIKubernetesResourceObservation{{Kind: "Pod", Name: "foreign", Namespace: "other", Owned: false}}}
	if err := o.Validate(); err == nil {
		t.Fatal("foreign resource must be rejected")
	}
}

func TestAIKubernetesPlanRequiresVerificationForMutation(t *testing.T) {
	now := time.Now().UTC()
	p := AIKubernetesAgentPlan{SchemaVersion: AIKubernetesAgentSchemaVersion, PlanID: "plan", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", ContextHash: "hash", GeneratedAt: now, Evidence: []AIEvidenceReference{{TenantID: "tenant-a", SourceID: "event:1"}}, Observation: AIKubernetesObservation{TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", Namespace: "env-a", NamespaceValid: true, OwnershipValid: true}, Proposals: []AIKubernetesRepairProposal{{Kind: AIKubernetesRestartOwnedWorkload, EnvironmentID: "env-a", Namespace: "env-a", Tool: "restart_owned_workload", CompensationGuidance: "restore previous rollout", PreVerification: []string{"ownership"}, PostVerification: nil, ApprovalRequired: true}}}
	if err := p.Validate(); err == nil {
		t.Fatal("mutation without post-verification must be rejected")
	}
}
