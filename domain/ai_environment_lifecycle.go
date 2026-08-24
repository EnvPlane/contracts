package domain

import (
	"errors"
	"strings"
	"time"
)

const AIEnvironmentLifecycleSchemaVersion = "1"

type AIEnvironmentLifecycleAction string

const (
	AIEnvironmentCreate    AIEnvironmentLifecycleAction = "create"
	AIEnvironmentReconcile AIEnvironmentLifecycleAction = "reconcile"
	AIEnvironmentRepair    AIEnvironmentLifecycleAction = "repair"
	AIEnvironmentResize    AIEnvironmentLifecycleAction = "resize"
	AIEnvironmentExtendTTL AIEnvironmentLifecycleAction = "extend_ttl"
	AIEnvironmentCleanup   AIEnvironmentLifecycleAction = "cleanup"
)

type AIEnvironmentReadinessCheck struct {
	ID      string `json:"id"`
	Passed  bool   `json:"passed"`
	Summary string `json:"summary"`
}

type AIEnvironmentLifecycleRequest struct {
	SchemaVersion  string                       `json:"schemaVersion"`
	ID             string                       `json:"id"`
	IdempotencyKey string                       `json:"idempotencyKey"`
	TenantID       string                       `json:"tenantId"`
	ProjectID      string                       `json:"projectId"`
	EnvironmentID  string                       `json:"environmentId,omitempty"`
	ProposalID     string                       `json:"proposalId,omitempty"`
	Action         AIEnvironmentLifecycleAction `json:"action"`
	ContextHash    string                       `json:"contextHash"`
	RequestedBy    string                       `json:"requestedBy"`
	CreatedAt      time.Time                    `json:"createdAt"`
}

type AIEnvironmentLifecyclePlan struct {
	SchemaVersion    string                        `json:"schemaVersion"`
	Request          AIEnvironmentLifecycleRequest `json:"request"`
	Plan             AIPlan                        `json:"plan"`
	Readiness        []AIEnvironmentReadinessCheck `json:"readiness"`
	Ready            bool                          `json:"ready"`
	ApprovalRequired bool                          `json:"approvalRequired"`
	ApprovalReason   string                        `json:"approvalReason,omitempty"`
}

type AIEnvironmentLifecycleReport struct {
	SchemaVersion   string       `json:"schemaVersion"`
	PlanID          string       `json:"planId"`
	TenantID        string       `json:"tenantId"`
	ProjectID       string       `json:"projectId"`
	EnvironmentID   string       `json:"environmentId,omitempty"`
	Status          AIPlanStatus `json:"status"`
	Summary         string       `json:"summary"`
	Resources       []string     `json:"resources,omitempty"`
	URL             string       `json:"url,omitempty"`
	Revision        string       `json:"revision,omitempty"`
	UnresolvedRisks []string     `json:"unresolvedRisks,omitempty"`
	GeneratedAt     time.Time    `json:"generatedAt"`
}

func (a AIEnvironmentLifecycleAction) valid() bool {
	switch a {
	case AIEnvironmentCreate, AIEnvironmentReconcile, AIEnvironmentRepair, AIEnvironmentResize, AIEnvironmentExtendTTL, AIEnvironmentCleanup:
		return true
	default:
		return false
	}
}

func (a AIEnvironmentLifecycleAction) RequiresApproval() bool {
	switch a {
	case AIEnvironmentCreate, AIEnvironmentRepair, AIEnvironmentResize, AIEnvironmentCleanup:
		return true
	default:
		return false
	}
}

func (r AIEnvironmentLifecycleRequest) Validate() error {
	if r.SchemaVersion != AIEnvironmentLifecycleSchemaVersion || strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.IdempotencyKey) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.ContextHash) == "" || strings.TrimSpace(r.RequestedBy) == "" || r.CreatedAt.IsZero() || !r.Action.valid() {
		return errors.New("AI environment lifecycle request is invalid")
	}
	if r.Action == AIEnvironmentCreate && strings.TrimSpace(r.ProposalID) == "" {
		return errors.New("environment creation requires a typed proposal")
	}
	if r.Action != AIEnvironmentCreate && strings.TrimSpace(r.EnvironmentID) == "" {
		return errors.New("environment lifecycle action requires an environment")
	}
	return nil
}

func (p AIEnvironmentLifecyclePlan) Validate() error {
	if p.SchemaVersion != AIEnvironmentLifecycleSchemaVersion || len(p.Readiness) == 0 || p.Ready != allReadinessPassed(p.Readiness) {
		return errors.New("AI environment lifecycle plan readiness is invalid")
	}
	if err := p.Request.Validate(); err != nil {
		return err
	}
	if p.ApprovalRequired != p.Request.Action.RequiresApproval() {
		return errors.New("AI environment lifecycle approval classification is invalid")
	}
	return p.Plan.Validate()
}

var (
	ErrAIEnvironmentLifecycleNotReady = errors.New("AI environment lifecycle plan is not ready")
	ErrAIEnvironmentLifecycleApproval = errors.New("AI environment lifecycle approval is required")
)

func (p AIEnvironmentLifecyclePlan) ValidateForExecution(approved bool) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.Ready {
		return ErrAIEnvironmentLifecycleNotReady
	}
	if p.ApprovalRequired && !approved {
		return ErrAIEnvironmentLifecycleApproval
	}
	return p.Plan.ValidateForExecution(p.Request.ContextHash)
}

func allReadinessPassed(checks []AIEnvironmentReadinessCheck) bool {
	if len(checks) == 0 {
		return false
	}
	for _, check := range checks {
		if strings.TrimSpace(check.ID) == "" || strings.TrimSpace(check.Summary) == "" || len(check.Summary) > 512 || !check.Passed {
			return false
		}
	}
	return true
}

func (r AIEnvironmentLifecycleReport) Validate() error {
	if r.SchemaVersion != AIEnvironmentLifecycleSchemaVersion || strings.TrimSpace(r.PlanID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.Summary) == "" || len(r.Summary) > 4000 || r.GeneratedAt.IsZero() {
		return errors.New("AI environment lifecycle report is invalid")
	}
	if r.Status != AIPlanSucceeded && r.Status != AIPlanFailed && r.Status != AIPlanCanceled && r.Status != AIPlanPaused {
		return errors.New("AI environment lifecycle report must be terminal or paused")
	}
	if len(r.Resources) > 128 || len(r.UnresolvedRisks) > 32 || len(r.URL) > 2048 || len(r.Revision) > 256 {
		return errors.New("AI environment lifecycle report is unbounded")
	}
	return nil
}
