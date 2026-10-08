package domain

func firstNonEmptyAssistantScope(scope string) string {
	if scope == "" {
		return "project"
	}
	return scope
}

// A tenant run is explicitly typed and never borrows a project identity.
func (r AIRun) TenantAssistantScope() bool {
	return r.ProjectID == "" && r.SubjectType == "tenant_assistant_snapshot" && (r.Purpose == "release.engineering" || r.Purpose == "approved_actions")
}
