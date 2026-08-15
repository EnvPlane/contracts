package domain

import (
	"fmt"
	"time"
)

// Job is the shared lifecycle envelope used by control-plane and runner.
// Event stays generic because SCM-specific payloads belong to SCM contracts.
type Job struct {
	TenantID      string                   `json:"tenant_id,omitempty"`
	ID            string                   `json:"id"`
	Type          JobType                  `json:"type"`
	Status        JobStatus                `json:"status"`
	EnvironmentID string                   `json:"environmentId"`
	DedupKey      string                   `json:"dedupKey,omitempty"`
	Event         any                      `json:"event"`
	Request       CreateEnvironmentRequest `json:"request"`
	Result        *Environment             `json:"result,omitempty"`
	Error         string                   `json:"error,omitempty"`
	Attempts      int                      `json:"attempts"`
	MaxAttempts   int                      `json:"maxAttempts"`
	NextRunAt     *time.Time               `json:"nextRunAt,omitempty"`
	CreatedAt     time.Time                `json:"createdAt"`
	StartedAt     *time.Time               `json:"startedAt,omitempty"`
	CompletedAt   *time.Time               `json:"completedAt,omitempty"`
}

// JobType is the cross-service operation vocabulary.
type JobType string

const (
	JobTypeCreateEnvironment JobType = "create_environment"
	JobTypeDeleteEnvironment JobType = "delete_environment"
)

// JobStatus is the durable lifecycle vocabulary shared by control-plane and
// runner. Unknown values must be rejected at contract boundaries.
type JobStatus string

const (
	JobStatusQueued          JobStatus = "queued"
	JobStatusRunning         JobStatus = "running"
	JobStatusSucceeded       JobStatus = "succeeded"
	JobStatusFailed          JobStatus = "failed"
	JobStatusIgnored         JobStatus = "ignored"
	JobStatusApprovalPending JobStatus = "approval_pending"
)

func AllJobStatuses() []JobStatus {
	return []JobStatus{JobStatusQueued, JobStatusRunning, JobStatusSucceeded, JobStatusFailed, JobStatusIgnored, JobStatusApprovalPending}
}

func ParseJobStatus(value string) (JobStatus, error) {
	status := JobStatus(value)
	for _, allowed := range AllJobStatuses() {
		if status == allowed {
			return status, nil
		}
	}
	return "", fmt.Errorf("unknown job status %q", value)
}
