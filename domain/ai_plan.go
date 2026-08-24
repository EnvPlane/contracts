package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const AIPlanSchemaVersion = "1"

var (
	ErrAIPlanNotExecutable = errors.New("AI plan is not approved for execution")
	ErrAIPlanIncompatible  = errors.New("AI plan schema is incompatible")
	ErrAIPlanStale         = errors.New("AI plan context is stale")
	ErrAIPlanTerminal      = errors.New("AI plan is terminal")
)

type AIPlanStatus string

const (
	AIPlanProposed  AIPlanStatus = "proposed"
	AIPlanValidated AIPlanStatus = "validated"
	AIPlanApproved  AIPlanStatus = "approved"
	AIPlanExecuting AIPlanStatus = "executing"
	AIPlanSucceeded AIPlanStatus = "succeeded"
	AIPlanFailed    AIPlanStatus = "failed"
	AIPlanCanceled  AIPlanStatus = "canceled"
	AIPlanPaused    AIPlanStatus = "paused"
)

type AIStepStatus string

const (
	AIStepPending   AIStepStatus = "pending"
	AIStepExecuting AIStepStatus = "executing"
	AIStepSucceeded AIStepStatus = "succeeded"
	AIStepFailed    AIStepStatus = "failed"
	AIStepSkipped   AIStepStatus = "skipped"
)

type AIToolCall struct {
	SchemaVersion  string            `json:"schemaVersion"`
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Arguments      map[string]string `json:"arguments"`
	IdempotencyKey string            `json:"idempotencyKey"`
	ContextHash    string            `json:"contextHash"`
}

type AIVerification struct {
	SchemaVersion string `json:"schemaVersion"`
	ID            string `json:"id"`
	Check         string `json:"check"`
	Expected      string `json:"expected"`
	ReadOnly      bool   `json:"readOnly"`
}

type AICompensation struct {
	SchemaVersion string     `json:"schemaVersion"`
	ID            string     `json:"id"`
	Tool          AIToolCall `json:"tool"`
	Reason        string     `json:"reason"`
}

type AIStep struct {
	SchemaVersion  string          `json:"schemaVersion"`
	ID             string          `json:"id"`
	Sequence       int             `json:"sequence"`
	Status         AIStepStatus    `json:"status"`
	ModelRationale string          `json:"modelRationale,omitempty"`
	Tool           AIToolCall      `json:"tool"`
	Verification   AIVerification  `json:"verification"`
	Compensation   *AICompensation `json:"compensation,omitempty"`
}

type AIFinalReport struct {
	SchemaVersion   string               `json:"schemaVersion"`
	Status          AIPlanStatus         `json:"status"`
	Summary         string               `json:"summary"`
	Verified        bool                 `json:"verified"`
	HumanEscalation bool                 `json:"humanEscalation"`
	ErrorCategory   AIProviderErrorClass `json:"errorCategory,omitempty"`
}

type AIPlan struct {
	SchemaVersion  string         `json:"schemaVersion"`
	ID             string         `json:"id"`
	IdempotencyKey string         `json:"idempotencyKey"`
	TenantID       string         `json:"tenantId"`
	ProjectID      string         `json:"projectId"`
	SubjectType    string         `json:"subjectType"`
	SubjectID      string         `json:"subjectId"`
	Purpose        string         `json:"purpose"`
	Action         string         `json:"action"`
	ContextHash    string         `json:"contextHash"`
	ModelNarrative string         `json:"modelNarrative"`
	Status         AIPlanStatus   `json:"status"`
	Steps          []AIStep       `json:"steps"`
	FinalReport    *AIFinalReport `json:"finalReport,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

func isAIPlanTerminal(status AIPlanStatus) bool {
	return status == AIPlanSucceeded || status == AIPlanFailed || status == AIPlanCanceled
}

func validateAIPlanStatus(status AIPlanStatus) error {
	switch status {
	case AIPlanProposed, AIPlanValidated, AIPlanApproved, AIPlanExecuting, AIPlanSucceeded, AIPlanFailed, AIPlanCanceled, AIPlanPaused:
		return nil
	default:
		return fmt.Errorf("unsupported AI plan status %q", status)
	}
}

func validateAIStepStatus(status AIStepStatus) error {
	switch status {
	case AIStepPending, AIStepExecuting, AIStepSucceeded, AIStepFailed, AIStepSkipped:
		return nil
	default:
		return fmt.Errorf("unsupported AI step status %q", status)
	}
}

func (t AIToolCall) Validate() error {
	if t.SchemaVersion != AIPlanSchemaVersion || strings.TrimSpace(t.ID) == "" || strings.TrimSpace(t.IdempotencyKey) == "" || strings.TrimSpace(t.ContextHash) == "" {
		return errors.New("AI tool call identity, idempotency key, and context hash are required")
	}
	allowed := map[string][]string{
		"read_environment_status":     {"environmentId"},
		"refresh_environment_status":  {"environmentId"},
		"retry_job":                   {"jobId"},
		"retry_gitops_reconciliation": {"projectId"},
		"apply_gitops_proposal":       {"proposalId"},
		"apply_typed_configuration":   {"proposalId"},
		"retry_kubernetes_rollout":    {"environmentId"},
		"refresh_kubernetes_scan":     {"environmentId"},
		"restart_owned_workload":      {"environmentId", "resourceName"},
		"create_environment":          {"proposalId"},
		"reconcile_environment":       {"environmentId"},
		"repair_environment":          {"environmentId"},
		"resize_environment":          {"environmentId", "proposalId"},
		"extend_environment_ttl":      {"environmentId"},
		"cleanup_environment":         {"environmentId"},
	}
	required, ok := allowed[t.Name]
	if !ok {
		return fmt.Errorf("AI tool %q is not allowlisted", t.Name)
	}
	if len(t.Arguments) != len(required) {
		return errors.New("AI tool arguments do not match the allowlist")
	}
	for _, key := range required {
		if strings.TrimSpace(t.Arguments[key]) == "" {
			return fmt.Errorf("AI tool argument %q is required", key)
		}
	}
	for key := range t.Arguments {
		found := false
		for _, allowedKey := range required {
			if key == allowedKey {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("AI tool argument %q is not allowlisted", key)
		}
	}
	return nil
}

func (v AIVerification) Validate() error {
	if v.SchemaVersion != AIPlanSchemaVersion || strings.TrimSpace(v.ID) == "" || strings.TrimSpace(v.Check) == "" || strings.TrimSpace(v.Expected) == "" || !v.ReadOnly {
		return errors.New("AI verification must be versioned, named, bounded, and read-only")
	}
	return nil
}

func (c AICompensation) Validate() error {
	if c.SchemaVersion != AIPlanSchemaVersion || strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.Reason) == "" {
		return errors.New("AI compensation identity and reason are required")
	}
	return c.Tool.Validate()
}

func (s AIStep) Validate() error {
	if s.SchemaVersion != AIPlanSchemaVersion || strings.TrimSpace(s.ID) == "" || s.Sequence < 1 {
		return errors.New("AI step identity and sequence are required")
	}
	if err := validateAIStepStatus(s.Status); err != nil {
		return err
	}
	if len(s.ModelRationale) > 4000 {
		return errors.New("AI model rationale exceeds the bounded length")
	}
	if err := s.Tool.Validate(); err != nil {
		return err
	}
	if err := s.Verification.Validate(); err != nil {
		return err
	}
	if s.Compensation != nil {
		if err := s.Compensation.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r AIFinalReport) Validate() error {
	if r.SchemaVersion != AIPlanSchemaVersion || strings.TrimSpace(r.Summary) == "" || len(r.Summary) > 4000 {
		return errors.New("AI final report is invalid or unbounded")
	}
	if err := validateAIPlanStatus(r.Status); err != nil || (!isAIPlanTerminal(r.Status) && r.Status != AIPlanPaused) {
		return errors.New("AI final report must describe a terminal or paused plan state")
	}
	if r.ErrorCategory != "" {
		return (AIProviderError{Class: r.ErrorCategory}).Validate()
	}
	return nil
}

func (p AIPlan) Validate() error {
	if p.SchemaVersion != AIPlanSchemaVersion || strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.IdempotencyKey) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.SubjectType) == "" || strings.TrimSpace(p.SubjectID) == "" || strings.TrimSpace(p.Purpose) == "" || strings.TrimSpace(p.Action) == "" || strings.TrimSpace(p.ContextHash) == "" {
		return errors.New("AI plan identity, scope, action, and context hash are required")
	}
	if err := validateAIPlanStatus(p.Status); err != nil {
		return err
	}
	if len(p.ModelNarrative) > 8000 {
		return errors.New("AI model narrative exceeds the bounded length")
	}
	if len(p.Steps) == 0 || len(p.Steps) > 32 {
		return errors.New("AI plan must contain one to thirty-two steps")
	}
	for index, step := range p.Steps {
		if step.Sequence != index+1 {
			return errors.New("AI plan step sequence is not contiguous")
		}
		if err := step.Validate(); err != nil {
			return err
		}
		if step.Tool.ContextHash != p.ContextHash || (step.Compensation != nil && step.Compensation.Tool.ContextHash != p.ContextHash) {
			return errors.New("AI tool context hash does not match plan")
		}
	}
	if p.FinalReport != nil {
		if err := p.FinalReport.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p AIPlan) ValidateForExecution(contextHash string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.Status != AIPlanApproved {
		return ErrAIPlanNotExecutable
	}
	if strings.TrimSpace(contextHash) == "" || p.ContextHash != contextHash {
		return ErrAIPlanStale
	}
	return nil
}

func (p AIPlan) ValidateResumeCompatibility(currentSchemaVersion, contextHash string) error {
	if p.SchemaVersion != currentSchemaVersion || currentSchemaVersion != AIPlanSchemaVersion {
		return ErrAIPlanIncompatible
	}
	if isAIPlanTerminal(p.Status) {
		return ErrAIPlanTerminal
	}
	if strings.TrimSpace(contextHash) == "" || p.ContextHash != contextHash {
		return ErrAIPlanStale
	}
	return nil
}

func (p AIPlan) PauseForIncompatibility(reason string) AIPlan {
	p.Status = AIPlanPaused
	p.UpdatedAt = time.Now().UTC()
	p.FinalReport = &AIFinalReport{SchemaVersion: AIPlanSchemaVersion, Status: AIPlanPaused, Summary: reason, HumanEscalation: true}
	return p
}

func (p AIPlan) RoundTrip() (AIPlan, error) {
	encoded, err := json.Marshal(p)
	if err != nil {
		return AIPlan{}, err
	}
	var decoded AIPlan
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return AIPlan{}, err
	}
	return decoded, nil
}
