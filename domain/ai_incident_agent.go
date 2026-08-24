package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

const AIIncidentAgentSchemaVersion = "1"

type AIIncidentState string

const (
	AIIncidentOpen          AIIncidentState = "open"
	AIIncidentNeedsApproval AIIncidentState = "needs_approval"
	AIIncidentEscalated     AIIncidentState = "escalated"
	AIIncidentResolved      AIIncidentState = "resolved"
)

type AIIncidentImpact string

const (
	AIIncidentImpactUnknown       AIIncidentImpact = "unknown"
	AIIncidentImpactDegraded      AIIncidentImpact = "degraded"
	AIIncidentImpactPartialOutage AIIncidentImpact = "partial_outage"
	AIIncidentImpactOutage        AIIncidentImpact = "outage"
)

type AIIncidentBlastRadius string

const (
	AIIncidentBlastEnvironment AIIncidentBlastRadius = "environment"
	AIIncidentBlastProject     AIIncidentBlastRadius = "project"
	AIIncidentBlastTenant      AIIncidentBlastRadius = "tenant"
	AIIncidentBlastPlatform    AIIncidentBlastRadius = "platform"
)

type AIIncidentEvidence struct {
	ID            string              `json:"id"`
	Reference     AIEvidenceReference `json:"reference"`
	ProjectID     string              `json:"project_id"`
	EnvironmentID string              `json:"environment_id"`
	Check         string              `json:"check"`
	Signal        string              `json:"signal"`
	FailureCode   string              `json:"failure_code,omitempty"`
	ObservedAt    time.Time           `json:"observed_at"`
	Resolved      bool                `json:"resolved"`
}

type AIIncidentActionKind string

const (
	AIIncidentRefreshStatus   AIIncidentActionKind = "refresh_status"
	AIIncidentRetryReconcile  AIIncidentActionKind = "retry_reconcile"
	AIIncidentRestartWorkload AIIncidentActionKind = "restart_owned_workload"
	AIIncidentRollback        AIIncidentActionKind = "rollback"
	AIIncidentFailover        AIIncidentActionKind = "failover"
)

type AIIncidentAction struct {
	Kind             AIIncidentActionKind `json:"kind"`
	Target           string               `json:"target"`
	ApprovalRequired bool                 `json:"approval_required"`
	Preconditions    []string             `json:"preconditions"`
	Verification     []string             `json:"verification"`
	Compensation     string               `json:"compensation"`
	EvidenceIDs      []string             `json:"evidence_ids"`
}

type AIIncidentReport struct {
	Impact             AIIncidentImpact     `json:"impact"`
	SuspectedCauseCode string               `json:"suspected_cause_code"`
	Confidence         AIConfidence         `json:"confidence"`
	EvidenceTimeline   []AIIncidentEvidence `json:"evidence_timeline"`
	ActionsTaken       []AIIncidentAction   `json:"actions_taken,omitempty"`
	Outcomes           []string             `json:"outcomes,omitempty"`
	PreventionTasks    []string             `json:"prevention_tasks"`
}

type AIIncidentAgentRequest struct {
	ProjectID     string                `json:"project_id"`
	EnvironmentID string                `json:"environment_id"`
	WindowStart   time.Time             `json:"window_start"`
	WindowEnd     time.Time             `json:"window_end"`
	FailureCount  int                   `json:"failure_count"`
	BlastRadius   AIIncidentBlastRadius `json:"blast_radius"`
	Impact        AIIncidentImpact      `json:"impact"`
	Evidence      []AIIncidentEvidence  `json:"evidence"`
}

type AIIncidentAgentPlan struct {
	SchemaVersion      string                `json:"schema_version"`
	IncidentID         string                `json:"incident_id"`
	TenantID           string                `json:"tenant_id"`
	ProjectID          string                `json:"project_id"`
	EnvironmentID      string                `json:"environment_id"`
	ContextHash        string                `json:"context_hash"`
	State              AIIncidentState       `json:"state"`
	Impact             AIIncidentImpact      `json:"impact"`
	BlastRadius        AIIncidentBlastRadius `json:"blast_radius"`
	SuspectedCauseCode string                `json:"suspected_cause_code"`
	Confidence         AIConfidence          `json:"confidence"`
	EvidenceTimeline   []AIIncidentEvidence  `json:"evidence_timeline"`
	Actions            []AIIncidentAction    `json:"actions"`
	EscalationReason   string                `json:"escalation_reason,omitempty"`
	PreventionTasks    []string              `json:"prevention_tasks"`
	FailClosed         bool                  `json:"fail_closed"`
	GeneratedAt        time.Time             `json:"generated_at"`
}

func (e AIIncidentEvidence) Validate(tenantID, projectID, environmentID string) error {
	if strings.TrimSpace(e.ID) == "" || strings.TrimSpace(e.Check) == "" || strings.TrimSpace(e.Signal) == "" || e.ObservedAt.IsZero() {
		return errors.New("incident evidence identity is required")
	}
	if e.Reference.TenantID != tenantID || e.Reference.SourceType == "" || e.Reference.SourceID == "" {
		return errors.New("incident evidence is outside tenant scope")
	}
	if e.ProjectID != projectID || e.EnvironmentID != environmentID {
		return errors.New("incident evidence is outside target scope")
	}
	if len(e.Check) > 128 || len(e.Signal) > 128 || len(e.FailureCode) > 128 {
		return errors.New("incident evidence field limit exceeded")
	}
	return nil
}

func (r AIIncidentAgentRequest) Validate(tenantID string) error {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.EnvironmentID) == "" || r.WindowStart.IsZero() || r.WindowEnd.IsZero() || !r.WindowEnd.After(r.WindowStart) {
		return errors.New("incident request scope or window is invalid")
	}
	if r.WindowEnd.Sub(r.WindowStart) > 24*time.Hour || r.FailureCount < 0 || r.FailureCount > 100 {
		return errors.New("incident request bounds are invalid")
	}
	if r.BlastRadius == "" || r.Impact == "" || len(r.Evidence) == 0 || len(r.Evidence) > 256 {
		return errors.New("incident request evidence is invalid")
	}
	for _, evidence := range r.Evidence {
		if err := evidence.Validate(tenantID, r.ProjectID, r.EnvironmentID); err != nil {
			return err
		}
	}
	return nil
}

func (p AIIncidentAgentPlan) Validate() error {
	if p.SchemaVersion != AIIncidentAgentSchemaVersion || strings.TrimSpace(p.IncidentID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.EnvironmentID) == "" || strings.TrimSpace(p.ContextHash) == "" || p.GeneratedAt.IsZero() || p.State == "" || p.Impact == "" || p.BlastRadius == "" || len(p.EvidenceTimeline) == 0 {
		return errors.New("incident plan identity is invalid")
	}
	for _, evidence := range p.EvidenceTimeline {
		if err := evidence.Validate(p.TenantID, p.ProjectID, p.EnvironmentID); err != nil {
			return err
		}
	}
	if p.Confidence.EvidenceCount != len(p.EvidenceTimeline) || p.Confidence.Score < 0 || p.Confidence.Score > 1 {
		return errors.New("incident confidence is invalid")
	}
	if p.FailClosed && strings.TrimSpace(p.EscalationReason) == "" {
		return errors.New("fail-closed incident plan needs an escalation reason")
	}
	for _, action := range p.Actions {
		if strings.TrimSpace(action.Target) == "" || len(action.Preconditions) == 0 || len(action.Verification) == 0 || strings.TrimSpace(action.Compensation) == "" || len(action.EvidenceIDs) == 0 {
			return errors.New("incident action lacks safety bounds")
		}
		if (action.Kind == AIIncidentRollback || action.Kind == AIIncidentFailover || action.Kind == AIIncidentRestartWorkload) && !action.ApprovalRequired {
			return errors.New("incident high-risk action requires approval")
		}
	}
	return nil
}

func SortIncidentEvidence(evidence []AIIncidentEvidence) []AIIncidentEvidence {
	result := append([]AIIncidentEvidence(nil), evidence...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ObservedAt.Equal(result[j].ObservedAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].ObservedAt.Before(result[j].ObservedAt)
	})
	return result
}
