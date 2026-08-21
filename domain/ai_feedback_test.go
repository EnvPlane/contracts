package domain

import (
	"testing"
	"time"
)

func TestAIFeedbackValidatesReasonAndBoundedComment(t *testing.T) {
	feedback := AIFeedback{SchemaVersion: AIFeedbackSchemaVersion, ID: "feedback-1", RunID: "run-1", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", Helpful: false, Reason: AIFeedbackReasonMissingEvidence, CreatedAt: time.Now().UTC()}
	if err := feedback.Validate(); err != nil { t.Fatal(err) }
	feedback.Comment = string(make([]rune, 501))
	if err := feedback.Validate(); err == nil { t.Fatal("unbounded comment accepted") }
}
