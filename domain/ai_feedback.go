package domain

import (
	"errors"
	"strings"
	"time"
)

const AIFeedbackSchemaVersion = "1"

type AIFeedbackReason string

const (
	AIFeedbackReasonInaccurate      AIFeedbackReason = "inaccurate"
	AIFeedbackReasonMissingEvidence AIFeedbackReason = "missing_evidence"
	AIFeedbackReasonTooVague        AIFeedbackReason = "too_vague"
	AIFeedbackReasonNotActionable   AIFeedbackReason = "not_actionable"
	AIFeedbackReasonOther           AIFeedbackReason = "other"
)

type AIFeedback struct {
	SchemaVersion  string           `json:"schemaVersion"`
	ID             string           `json:"id"`
	RunID          string           `json:"runId"`
	TenantID       string           `json:"tenantId"`
	ProjectID      string           `json:"projectId"`
	EnvironmentID  string           `json:"environmentId"`
	Helpful        bool             `json:"helpful"`
	Reason         AIFeedbackReason `json:"reason"`
	Comment        string           `json:"comment,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
}

func (f AIFeedback) Validate() error {
	if f.SchemaVersion != AIFeedbackSchemaVersion || strings.TrimSpace(f.ID) == "" || strings.TrimSpace(f.RunID) == "" || strings.TrimSpace(f.TenantID) == "" || strings.TrimSpace(f.ProjectID) == "" || strings.TrimSpace(f.EnvironmentID) == "" || f.CreatedAt.IsZero() {
		return errors.New("AI feedback identity is invalid")
	}
	switch f.Reason {
	case AIFeedbackReasonInaccurate, AIFeedbackReasonMissingEvidence, AIFeedbackReasonTooVague, AIFeedbackReasonNotActionable, AIFeedbackReasonOther:
	default:
		return errors.New("AI feedback reason is invalid")
	}
	if !f.Helpful && f.Reason == "" {
		return errors.New("negative AI feedback requires a reason")
	}
	if len([]rune(f.Comment)) > 500 || strings.ContainsAny(f.Comment, "\r\n\x00") {
		return errors.New("AI feedback comment is invalid")
	}
	return nil
}
