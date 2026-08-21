package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const AIContractVersion = "1"

type AIRequestKind string

const (
	AIRequestKindDiagnosis AIRequestKind = "diagnosis"
	AIRequestKindBootstrap AIRequestKind = "bootstrap"
)

type AIRunStatus string

const (
	AIRunStatusQueued    AIRunStatus = "queued"
	AIRunStatusRunning   AIRunStatus = "running"
	AIRunStatusSucceeded AIRunStatus = "succeeded"
	AIRunStatusFailed    AIRunStatus = "failed"
	AIRunStatusCanceled  AIRunStatus = "canceled"
)

type AIDiagnosisOutcome string

const (
	AIDiagnosisOutcomeDiagnosis            AIDiagnosisOutcome = "diagnosis"
	AIDiagnosisOutcomeInsufficientEvidence AIDiagnosisOutcome = "insufficient_evidence"
)

type AIConfidenceLevel string

const (
	AIConfidenceHigh    AIConfidenceLevel = "high"
	AIConfidenceMedium  AIConfidenceLevel = "medium"
	AIConfidenceLow     AIConfidenceLevel = "low"
	AIConfidenceUnknown AIConfidenceLevel = "unknown"
)

type AIProviderErrorClass string

const (
	AIProviderErrorTransient           AIProviderErrorClass = "transient"
	AIProviderErrorRateLimited         AIProviderErrorClass = "rate_limited"
	AIProviderErrorInvalidRequest      AIProviderErrorClass = "invalid_request"
	AIProviderErrorAuthentication      AIProviderErrorClass = "authentication"
	AIProviderErrorUnauthorized        AIProviderErrorClass = "unauthorized"
	AIProviderErrorUnavailable         AIProviderErrorClass = "unavailable"
	AIProviderErrorProviderUnavailable AIProviderErrorClass = "provider_unavailable"
	AIProviderErrorTimeout             AIProviderErrorClass = "timeout"
	AIProviderErrorInvalidResponse     AIProviderErrorClass = "invalid_response"
	AIProviderErrorPolicyBlocked       AIProviderErrorClass = "policy_blocked"
	AIProviderErrorUnknown             AIProviderErrorClass = "unknown"
)

type AIEvidenceReference struct {
	SourceType string    `json:"sourceType"`
	SourceID   string    `json:"sourceId"`
	TenantID   string    `json:"tenantId"`
	ObservedAt time.Time `json:"observedAt,omitempty"`
	Digest     string    `json:"digest,omitempty"`
}

func (r AIEvidenceReference) Validate() error {
	if strings.TrimSpace(r.SourceType) == "" || strings.TrimSpace(r.SourceID) == "" || strings.TrimSpace(r.TenantID) == "" {
		return errors.New("AI evidence requires source type, stable source ID, and tenant ID")
	}
	return nil
}

type AIRequest struct {
	SchemaVersion string                `json:"schemaVersion"`
	RequestID     string                `json:"requestId"`
	TenantID      string                `json:"tenantId"`
	Kind          AIRequestKind         `json:"kind"`
	SubjectType   string                `json:"subjectType"`
	SubjectID     string                `json:"subjectId"`
	Evidence      []AIEvidenceReference `json:"evidence"`
	RequestedAt   time.Time             `json:"requestedAt"`
}

func (r AIRequest) Validate() error {
	if r.SchemaVersion != AIContractVersion || strings.TrimSpace(r.RequestID) == "" || strings.TrimSpace(r.TenantID) == "" {
		return errors.New("AI request identity or schema version is invalid")
	}
	if (r.Kind != AIRequestKindDiagnosis && r.Kind != AIRequestKindBootstrap) || strings.TrimSpace(r.SubjectType) == "" || strings.TrimSpace(r.SubjectID) == "" {
		return errors.New("AI request kind or subject is invalid")
	}
	for _, evidence := range r.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
		if evidence.TenantID != r.TenantID {
			return errors.New("AI evidence tenant does not match request tenant")
		}
	}
	return nil
}

type AIRunStatusRecord struct {
	SchemaVersion string           `json:"schemaVersion"`
	RunID         string           `json:"runId"`
	RequestID     string           `json:"requestId"`
	TenantID      string           `json:"tenantId"`
	Status        AIRunStatus      `json:"status"`
	UpdatedAt     time.Time        `json:"updatedAt"`
	ProviderError *AIProviderError `json:"providerError,omitempty"`
}

func (s AIRunStatusRecord) Validate() error {
	if s.SchemaVersion != AIContractVersion || strings.TrimSpace(s.RunID) == "" || strings.TrimSpace(s.RequestID) == "" || strings.TrimSpace(s.TenantID) == "" {
		return errors.New("AI run status identity or schema version is invalid")
	}
	switch s.Status {
	case AIRunStatusQueued, AIRunStatusRunning, AIRunStatusSucceeded, AIRunStatusFailed, AIRunStatusCanceled:
		return nil
	default:
		return fmt.Errorf("unsupported AI run status %q", s.Status)
	}
}

type AIConfidence struct {
	Score         float64           `json:"score"`
	Level         AIConfidenceLevel `json:"level"`
	EvidenceCount int               `json:"evidenceCount"`
}

type AIProviderError struct {
	Class     AIProviderErrorClass `json:"class"`
	Code      string               `json:"code,omitempty"`
	Message   string               `json:"message,omitempty"`
	Retryable bool                 `json:"retryable"`
}

func (e AIProviderError) Validate() error {
	switch e.Class {
	case AIProviderErrorTransient, AIProviderErrorRateLimited, AIProviderErrorInvalidRequest,
		AIProviderErrorAuthentication, AIProviderErrorUnauthorized, AIProviderErrorUnavailable,
		AIProviderErrorProviderUnavailable, AIProviderErrorTimeout, AIProviderErrorInvalidResponse,
		AIProviderErrorPolicyBlocked, AIProviderErrorUnknown:
		return nil
	default:
		return fmt.Errorf("unsupported AI provider error class %q", e.Class)
	}
}

type AIDiagnosisResult struct {
	SchemaVersion  string                 `json:"schemaVersion"`
	RequestID      string                 `json:"requestId"`
	TenantID       string                 `json:"tenantId"`
	Outcome        AIDiagnosisOutcome     `json:"outcome"`
	Summary        string                 `json:"summary,omitempty"`
	Observed       []string               `json:"observed,omitempty"`
	Confidence     AIConfidence           `json:"confidence"`
	Evidence       []AIEvidenceReference  `json:"evidence"`
	LikelyCauses   []AIDiagnosisCause     `json:"likelyCauses,omitempty"`
	SafeChecks     []AIDiagnosisSafeCheck `json:"safeChecks,omitempty"`
	UserChecks     []AIDiagnosisSafeCheck `json:"userChecks,omitempty"`
	PlatformChecks []AIDiagnosisSafeCheck `json:"platformChecks,omitempty"`
	ProviderError  *AIProviderError       `json:"providerError,omitempty"`
}

type AIDiagnosisCause struct {
	ID         string                `json:"id"`
	Summary    string                `json:"summary"`
	Evidence   []AIEvidenceReference `json:"evidence,omitempty"`
	Hypothesis bool                  `json:"hypothesis"`
}

type AIDiagnosisSafeCheck struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	ReadOnly bool   `json:"readOnly"`
}

func (r AIDiagnosisResult) Validate() error {
	if r.SchemaVersion != AIContractVersion || strings.TrimSpace(r.RequestID) == "" || strings.TrimSpace(r.TenantID) == "" {
		return errors.New("AI diagnosis identity or schema version is invalid")
	}
	if r.Confidence.Score < 0 || r.Confidence.Score > 1 || math.IsNaN(r.Confidence.Score) || math.IsInf(r.Confidence.Score, 0) {
		return errors.New("AI confidence score must be between zero and one")
	}
	if r.Confidence.EvidenceCount != len(r.Evidence) {
		return errors.New("AI confidence evidence count does not match evidence")
	}
	for _, evidence := range r.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
		if evidence.TenantID != r.TenantID {
			return errors.New("AI diagnosis evidence tenant does not match result tenant")
		}
	}
	if len(r.Observed) > 20 || len(r.LikelyCauses) > 20 || len(r.SafeChecks) > 20 || len(r.UserChecks) > 20 || len(r.PlatformChecks) > 20 {
		return errors.New("AI diagnosis result exceeds bounded cause or check count")
	}
	for _, cause := range r.LikelyCauses {
		if strings.TrimSpace(cause.ID) == "" || strings.TrimSpace(cause.Summary) == "" {
			return errors.New("AI diagnosis cause requires an ID and summary")
		}
		if !cause.Hypothesis && len(cause.Evidence) == 0 {
			return errors.New("unsupported AI cause must be marked as a hypothesis")
		}
		for _, evidence := range cause.Evidence {
			if err := evidence.Validate(); err != nil || evidence.TenantID != r.TenantID {
				return errors.New("AI diagnosis cause evidence tenant is invalid")
			}
		}
	}
	allChecks := append(append(append([]AIDiagnosisSafeCheck{}, r.SafeChecks...), r.UserChecks...), r.PlatformChecks...)
	for _, check := range allChecks {
		if strings.TrimSpace(check.ID) == "" || strings.TrimSpace(check.Summary) == "" || !check.ReadOnly || strings.Contains(strings.ToLower(check.Summary), "kubectl") || strings.Contains(strings.ToLower(check.Summary), "helm") || strings.Contains(strings.ToLower(check.Summary), "shell") {
			return errors.New("AI safe checks must be named, read-only, and non-executable")
		}
	}
	if r.ProviderError != nil {
		if err := r.ProviderError.Validate(); err != nil {
			return err
		}
	}
	switch r.Outcome {
	case AIDiagnosisOutcomeDiagnosis:
		if len(r.Evidence) == 0 || r.Confidence.Level == AIConfidenceUnknown {
			return errors.New("AI diagnosis requires evidence and known confidence")
		}
	case AIDiagnosisOutcomeInsufficientEvidence:
		if len(r.Evidence) != 0 || r.Confidence.EvidenceCount != 0 || r.Confidence.Score != 0 || r.Confidence.Level != AIConfidenceUnknown {
			return errors.New("insufficient evidence must not claim confidence")
		}
	default:
		return fmt.Errorf("unsupported AI diagnosis outcome %q", r.Outcome)
	}
	return nil
}
