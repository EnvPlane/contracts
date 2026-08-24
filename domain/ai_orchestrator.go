package domain

import (
	"errors"
	"strings"
	"time"
)

const AIOrchestratorSchemaVersion = "1"

type AIOrchestratorStatus string

const (
	AIOrchestratorObserving   AIOrchestratorStatus = "observe"
	AIOrchestratorPlanning    AIOrchestratorStatus = "plan"
	AIOrchestratorAuthorizing AIOrchestratorStatus = "authorize"
	AIOrchestratorExecuting   AIOrchestratorStatus = "execute"
	AIOrchestratorVerifying   AIOrchestratorStatus = "verify"
	AIOrchestratorCompensate  AIOrchestratorStatus = "compensate"
	AIOrchestratorReported    AIOrchestratorStatus = "report"
	AIOrchestratorSucceeded   AIOrchestratorStatus = "succeeded"
	AIOrchestratorFailed      AIOrchestratorStatus = "failed"
	AIOrchestratorCanceled    AIOrchestratorStatus = "canceled"
	AIOrchestratorPaused      AIOrchestratorStatus = "paused"
)

type AIExecutionLease struct {
	Owner     string    `json:"owner"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type AIStepAttempt struct {
	StepID         string     `json:"stepId"`
	AttemptID      string     `json:"attemptId"`
	IdempotencyKey string     `json:"idempotencyKey"`
	StartedAt      time.Time  `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Status         string     `json:"status"`
	ErrorCategory  string     `json:"errorCategory,omitempty"`
}

type AIExecutionCheckpoint struct {
	Phase           AIOrchestratorStatus `json:"phase"`
	StepIndex       int                  `json:"stepIndex"`
	PlanContextHash string               `json:"planContextHash"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

type AIOrchestratorLimits struct {
	MaxIterations      int   `json:"maxIterations"`
	MaxToolCalls       int   `json:"maxToolCalls"`
	MaxTokens          int64 `json:"maxTokens"`
	MaxCostMicros      int64 `json:"maxCostMicros"`
	MaxWallTimeSeconds int   `json:"maxWallTimeSeconds"`
}

type AIOrchestratorRun struct {
	SchemaVersion  string                `json:"schemaVersion"`
	ID             string                `json:"id"`
	IdempotencyKey string                `json:"idempotencyKey"`
	TenantID       string                `json:"tenantId"`
	ProjectID      string                `json:"projectId"`
	Purpose        string                `json:"purpose"`
	Status         AIOrchestratorStatus  `json:"status"`
	Plan           AIPlan                `json:"plan"`
	Lease          *AIExecutionLease     `json:"lease,omitempty"`
	Attempts       []AIStepAttempt       `json:"attempts"`
	Checkpoint     AIExecutionCheckpoint `json:"checkpoint"`
	Limits         AIOrchestratorLimits  `json:"limits"`
	TokenCount     int64                 `json:"tokenCount"`
	CostMicros     int64                 `json:"costMicros"`
	ToolCalls      int                   `json:"toolCalls"`
	Iterations     int                   `json:"iterations"`
	Deadline       time.Time             `json:"deadline"`
	Version        int64                 `json:"version"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

func (r AIOrchestratorRun) Validate() error {
	if r.SchemaVersion != AIOrchestratorSchemaVersion || strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.IdempotencyKey) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.Purpose) == "" {
		return errors.New("AI orchestrator identity and tenant scope are required")
	}
	switch r.Status {
	case AIOrchestratorObserving, AIOrchestratorPlanning, AIOrchestratorAuthorizing, AIOrchestratorExecuting, AIOrchestratorVerifying, AIOrchestratorCompensate, AIOrchestratorReported, AIOrchestratorSucceeded, AIOrchestratorFailed, AIOrchestratorCanceled, AIOrchestratorPaused:
	default:
		return errors.New("AI orchestrator status is invalid")
	}
	if err := r.Plan.Validate(); err != nil {
		return err
	}
	if r.Version < 1 || r.TokenCount < 0 || r.CostMicros < 0 || r.ToolCalls < 0 || r.Iterations < 0 || r.Limits.MaxIterations <= 0 || r.Limits.MaxToolCalls <= 0 || r.Limits.MaxTokens <= 0 || r.Limits.MaxCostMicros < 0 || r.Limits.MaxWallTimeSeconds <= 0 || r.Deadline.IsZero() || r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() {
		return errors.New("AI orchestrator limits, counters, or timestamps are invalid")
	}
	if r.Lease != nil && (strings.TrimSpace(r.Lease.Owner) == "" || strings.TrimSpace(r.Lease.Token) == "" || r.Lease.ExpiresAt.IsZero()) {
		return errors.New("AI orchestrator lease is invalid")
	}
	return nil
}

func (r AIOrchestratorRun) Terminal() bool {
	return r.Status == AIOrchestratorSucceeded || r.Status == AIOrchestratorFailed || r.Status == AIOrchestratorCanceled
}
