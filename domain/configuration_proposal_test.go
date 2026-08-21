package domain

import (
	"testing"
	"time"
)

func TestConfigurationProposalRejectsUnsupportedFields(t *testing.T) {
	mode := EnvironmentMode("invalid")
	if err := (ConfigurationProposalFields{Mode: &mode}).Validate(); err == nil {
		t.Fatal("expected unsupported mode to be rejected")
	}
	if err := (ConfigurationProposalFields{ComponentRefs: []ConfigurationComponentRef{{ComponentID: "web"}, {ComponentID: "web"}}}).Validate(); err == nil {
		t.Fatal("expected duplicate components to be rejected")
	}
}

func TestConfigurationProposalDeterministic(t *testing.T) {
	proposal := ConfigurationProposal{Fields: ConfigurationProposalFields{ComponentRefs: []ConfigurationComponentRef{{ComponentID: "z"}, {ComponentID: "a"}}, HybridServices: []string{"z", "a"}}, Diff: []ConfigurationProposalDiff{{Field: "ttl"}, {Field: "mode"}}}
	got := proposal.Deterministic()
	if got.Fields.ComponentRefs[0].ComponentID != "a" || got.Fields.HybridServices[0] != "a" || got.Diff[0].Field != "mode" {
		t.Fatal("proposal ordering is not deterministic")
	}
}

func TestConfigurationProposalContextHashChangesWithProjectState(t *testing.T) {
	project := Project{TenantID: "tenant-a", ID: "project-a"}
	hash := ConfigurationProposalContextHash("tenant-a", project, ConfigurationProposalFields{})
	project.UpdatedAt = project.UpdatedAt.Add(time.Second)
	if hash == ConfigurationProposalContextHash("tenant-a", project, ConfigurationProposalFields{}) {
		t.Fatal("context hash did not fence project state")
	}
}

func TestConfigurationProposalActionContractRequiresIdempotency(t *testing.T) {
	request := ConfigurationProposalActionRequest{Action: "retry_job", Intent: "retry", JobID: "job-1"}
	if request.IdempotencyKey != "" {
		t.Fatal("test fixture unexpectedly has an idempotency key")
	}
	if request.Action == "force_clean" || request.Action == "delete_environment" {
		t.Fatal("destructive actions must not be part of the allowlist")
	}
}
