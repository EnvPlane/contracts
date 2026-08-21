package domain

import (
	"errors"
	"strings"
	"time"
)

const AIRunSchemaVersion = "1"

type AIRun struct {
	SchemaVersion        string             `json:"schemaVersion"`
	ID                   string             `json:"id"`
	IdempotencyKey       string             `json:"idempotencyKey"`
	TenantID             string             `json:"tenantId"`
	ProjectID            string             `json:"projectId"`
	Purpose              string             `json:"purpose"`
	Status               AIRunStatus        `json:"status"`
	Provider             string             `json:"provider"`
	Model                string             `json:"model"`
	PromptTemplateVersion string            `json:"promptTemplateVersion"`
	ContextHash          string             `json:"contextHash"`
	RequestedAt          time.Time          `json:"requestedAt"`
	StartedAt            *time.Time         `json:"startedAt,omitempty"`
	CompletedAt          *time.Time         `json:"completedAt,omitempty"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
	InputTokens          int64              `json:"inputTokens,omitempty"`
	OutputTokens         int64              `json:"outputTokens,omitempty"`
	LatencyMilliseconds  int64              `json:"latencyMilliseconds,omitempty"`
	ErrorCategory        AIProviderErrorClass `json:"errorCategory,omitempty"`
}

func (r AIRun) Validate() error {
	if r.SchemaVersion != AIRunSchemaVersion || strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.IdempotencyKey) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.Purpose) == "" || strings.TrimSpace(r.Provider) == "" || strings.TrimSpace(r.Model) == "" || strings.TrimSpace(r.PromptTemplateVersion) == "" || strings.TrimSpace(r.ContextHash) == "" {
		return errors.New("AI run identity and metadata are required")
	}
	if r.Status != AIRunStatusQueued && r.Status != AIRunStatusRunning && r.Status != AIRunStatusSucceeded && r.Status != AIRunStatusFailed && r.Status != AIRunStatusCanceled {
		return errors.New("AI run status is invalid")
	}
	if r.RequestedAt.IsZero() || r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() || r.InputTokens < 0 || r.OutputTokens < 0 || r.LatencyMilliseconds < 0 {
		return errors.New("AI run timestamps or counters are invalid")
	}
	if r.ErrorCategory != "" {
		if err := (AIProviderError{Class: r.ErrorCategory}).Validate(); err != nil { return err }
	}
	return nil
}

func ValidateAIRunTransition(from, to AIRunStatus) error {
	if from == to {
		if from == AIRunStatusSucceeded || from == AIRunStatusFailed || from == AIRunStatusCanceled { return nil }
		return nil
	}
	switch from {
	case AIRunStatusQueued:
		if to == AIRunStatusRunning || to == AIRunStatusFailed || to == AIRunStatusCanceled { return nil }
	case AIRunStatusRunning:
		if to == AIRunStatusSucceeded || to == AIRunStatusFailed || to == AIRunStatusCanceled { return nil }
	}
	return errors.New("AI run terminal or invalid status transition")
}
