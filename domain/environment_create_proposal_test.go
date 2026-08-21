package domain

import "testing"

func TestEnvironmentCreateProposalRejectsUnknownOrUnsafeFields(t *testing.T) {
	branch := "feature/demo"
	fields := EnvironmentCreateProposalFields{Branch: branch, Components: []ComponentOverride{{ComponentID: "api"}}}
	if err := fields.Validate(); err != nil {
		t.Fatal(err)
	}
	fields.Components = append(fields.Components, ComponentOverride{ComponentID: "api"})
	if err := fields.Validate(); err == nil {
		t.Fatal("duplicate component must be rejected")
	}
	fields = EnvironmentCreateProposalFields{Branch: branch, ResourceLimits: &ConfigurationResourceLimits{CPU: "250m"}}
	if err := fields.Validate(); err != nil {
		t.Fatal(err)
	}
	fields = EnvironmentCreateProposalFields{Branch: branch, ResourceLimits: &ConfigurationResourceLimits{CPU: "", Memory: ""}}
	if err := fields.Validate(); err == nil {
		t.Fatal("empty resource limits must be rejected")
	}
}

func TestEnvironmentCreateProposalContextChangesWhenProjectChanges(t *testing.T) {
	project := Project{TenantID: "tenant-a", ID: "project-a"}
	fields := EnvironmentCreateProposalFields{Branch: "main"}
	first := EnvironmentCreateProposalContextHash("tenant-a", project, fields)
	project.UpdatedAt = project.UpdatedAt.Add(1)
	if first == EnvironmentCreateProposalContextHash("tenant-a", project, fields) {
		t.Fatal("project update must invalidate proposal context")
	}
}
