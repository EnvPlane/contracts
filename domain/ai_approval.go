package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

const AIApprovalSchemaVersion = "1"

type AIApprovalMode string

const (
	AIApprovalNone        AIApprovalMode = "none"
	AIApprovalSingle      AIApprovalMode = "single"
	AIApprovalTwoPerson   AIApprovalMode = "two_person"
	AIApprovalMaintenance AIApprovalMode = "maintenance_window"
	AIApprovalProhibited  AIApprovalMode = "prohibited"
)

type AIApprovalState string

const (
	AIApprovalPending  AIApprovalState = "pending"
	AIApprovalApproved AIApprovalState = "approved"
	AIApprovalExpired  AIApprovalState = "expired"
	AIApprovalCanceled AIApprovalState = "canceled"
	AIApprovalDenied   AIApprovalState = "denied"
)

type AIRiskInput struct {
	Action           string `json:"action"`
	TargetType       string `json:"targetType"`
	EnvironmentClass string `json:"environmentClass"`
	BlastRadius      int    `json:"blastRadius"`
	CurrentHealth    string `json:"currentHealth"`
	Production       bool   `json:"production"`
	Delete           bool   `json:"delete"`
	Credential       bool   `json:"credential"`
	ClusterWide      bool   `json:"clusterWide"`
}

type AIRiskDecision struct {
	SchemaVersion string         `json:"schemaVersion"`
	Risk          AIRiskClass    `json:"risk"`
	Approval      AIApprovalMode `json:"approval"`
	Autonomous    bool           `json:"autonomous"`
	Prohibited    bool           `json:"prohibited"`
	Reason        string         `json:"reason"`
}

type AIApprovalRecord struct {
	SchemaVersion       string            `json:"schemaVersion"`
	ID                  string            `json:"id"`
	RunID               string            `json:"runId"`
	TenantID            string            `json:"tenantId"`
	ProjectID           string            `json:"projectId"`
	RequesterID         string            `json:"requesterId"`
	Action              string            `json:"action"`
	ActionArguments     map[string]string `json:"actionArguments"`
	ResourceID          string            `json:"resourceId"`
	PlanHash            string            `json:"planHash"`
	TargetStateHash     string            `json:"targetStateHash"`
	PolicyBundleVersion string            `json:"policyBundleVersion"`
	Risk                AIRiskDecision    `json:"risk"`
	ReviewedStepIDs     []string          `json:"reviewedStepIds"`
	Impact              string            `json:"impact"`
	Rollback            string            `json:"rollback"`
	State               AIApprovalState   `json:"state"`
	Approvers           []string          `json:"approvers"`
	MaintenanceStart    *time.Time        `json:"maintenanceStart,omitempty"`
	MaintenanceEnd      *time.Time        `json:"maintenanceEnd,omitempty"`
	ExpiresAt           time.Time         `json:"expiresAt"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

func (d AIRiskDecision) Validate() error {
	if d.SchemaVersion != AIApprovalSchemaVersion || strings.TrimSpace(d.Reason) == "" {
		return errors.New("AI risk decision is invalid")
	}
	switch d.Risk {
	case AIRiskLow, AIRiskMedium, AIRiskHigh, AIRiskCritical:
	default:
		return errors.New("AI risk class is invalid")
	}
	switch d.Approval {
	case AIApprovalNone, AIApprovalSingle, AIApprovalTwoPerson, AIApprovalMaintenance, AIApprovalProhibited:
	default:
		return errors.New("AI approval mode is invalid")
	}
	if d.Prohibited && d.Approval != AIApprovalProhibited {
		return errors.New("prohibited risk must use prohibited approval mode")
	}
	if d.Prohibited || d.Approval != AIApprovalNone {
		if d.Autonomous {
			return errors.New("approval-required risk cannot be autonomous")
		}
	}
	return nil
}

func ClassifyAIRisk(input AIRiskInput) AIRiskDecision {
	decision := AIRiskDecision{SchemaVersion: AIApprovalSchemaVersion, Risk: AIRiskLow, Approval: AIApprovalNone, Autonomous: true}
	switch {
	case input.Delete || input.Credential || input.ClusterWide:
		decision.Risk, decision.Approval, decision.Autonomous, decision.Prohibited = AIRiskCritical, AIApprovalProhibited, false, true
		decision.Reason = "delete, credential, and cluster-wide actions are prohibited for AI execution"
	case input.Production && input.BlastRadius > 1:
		decision.Risk, decision.Approval, decision.Autonomous = AIRiskHigh, AIApprovalTwoPerson, false
		decision.Reason = "production action has multi-target blast radius"
	case input.Production:
		decision.Risk, decision.Approval, decision.Autonomous = AIRiskHigh, AIApprovalMaintenance, false
		decision.Reason = "production action requires a maintenance window and approval"
	case input.BlastRadius > 10 || strings.EqualFold(input.EnvironmentClass, "critical"):
		decision.Risk, decision.Approval, decision.Autonomous = AIRiskHigh, AIApprovalTwoPerson, false
		decision.Reason = "large or critical target scope requires two-person approval"
	case input.BlastRadius > 1 || !strings.EqualFold(input.CurrentHealth, "healthy"):
		decision.Risk, decision.Approval, decision.Autonomous = AIRiskMedium, AIApprovalSingle, false
		decision.Reason = "non-trivial blast radius or unhealthy target requires approval"
	default:
		decision.Reason = "read-only or bounded single-target action"
	}
	return decision
}

func (r AIApprovalRecord) Validate(now time.Time) error {
	if r.SchemaVersion != AIApprovalSchemaVersion || strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.RunID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.RequesterID) == "" || strings.TrimSpace(r.Action) == "" || strings.TrimSpace(r.ResourceID) == "" || strings.TrimSpace(r.PlanHash) == "" || strings.TrimSpace(r.TargetStateHash) == "" || strings.TrimSpace(r.PolicyBundleVersion) == "" || strings.TrimSpace(r.Impact) == "" || strings.TrimSpace(r.Rollback) == "" {
		return errors.New("AI approval identity and review data are required")
	}
	if err := r.Risk.Validate(); err != nil {
		return err
	}
	if len(r.ReviewedStepIDs) == 0 {
		return errors.New("AI approval must enumerate reviewed steps")
	}
	if len(r.ActionArguments) > 32 || len(r.ReviewedStepIDs) > 128 || len(r.Approvers) > 2 {
		return errors.New("AI approval fields are unbounded")
	}
	steps := append([]string(nil), r.ReviewedStepIDs...)
	sort.Strings(steps)
	for i, step := range steps {
		if strings.TrimSpace(step) == "" || (i > 0 && step == steps[i-1]) {
			return errors.New("AI approval step list is invalid")
		}
	}
	if r.ExpiresAt.IsZero() || r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() || !r.ExpiresAt.After(r.CreatedAt) {
		return errors.New("AI approval timestamps are invalid")
	}
	if !now.IsZero() && r.ExpiresAt.Before(now.UTC()) && r.State == AIApprovalPending {
		return errors.New("AI approval is expired")
	}
	return nil
}
