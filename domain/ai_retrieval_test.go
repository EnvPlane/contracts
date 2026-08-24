package domain

import "testing"

func TestAIRetrievalQueryAndConclusionRequireAllowlistedEvidence(t *testing.T) {
	query := AIRetrievalQuery{SchemaVersion: AIRetrievalSchemaVersion, TenantID: "tenant-a", ProjectID: "project-a", SubjectType: "environment", SubjectID: "env-a", Sources: []string{AIRetrievalEvent}, MaxResults: 4}
	if err := query.Validate(); err != nil {
		t.Fatal(err)
	}
	conclusion := AIMaterialConclusion{SchemaVersion: AIRetrievalSchemaVersion, Summary: "event is relevant", EvidenceIDs: []string{"evidence-1"}}
	if err := conclusion.Validate(map[string]struct{}{"evidence-1": {}}); err != nil {
		t.Fatal(err)
	}
	if err := conclusion.Validate(map[string]struct{}{}); err == nil {
		t.Fatal("unavailable evidence accepted")
	}
}
