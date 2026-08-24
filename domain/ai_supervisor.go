package domain

import (
	"errors"
	"strings"
	"time"
)

const AISupervisorSchemaVersion = "1"

type AISupervisorAgentRole string

const (
	AISupervisorDiagnosis  AISupervisorAgentRole = "diagnosis"
	AISupervisorSecurity   AISupervisorAgentRole = "security_compliance"
	AISupervisorGitOps     AISupervisorAgentRole = "gitops"
	AISupervisorKubernetes AISupervisorAgentRole = "kubernetes"
	AISupervisorFinOps     AISupervisorAgentRole = "finops"
	AISupervisorRelease    AISupervisorAgentRole = "release_engineering"
)

type AISupervisorBudget struct {
	MaxDepth           int   `json:"max_depth"`
	MaxFanOut          int   `json:"max_fan_out"`
	MaxParallel        int   `json:"max_parallel"`
	MaxCostMicros      int64 `json:"max_cost_micros"`
	MaxWallTimeSeconds int   `json:"max_wall_time_seconds"`
}

type AISupervisorSubgoal struct {
	ID            string                `json:"id"`
	ParentID      string                `json:"parent_id,omitempty"`
	GoalCode      string                `json:"goal_code"`
	Role          AISupervisorAgentRole `json:"role"`
	TenantID      string                `json:"tenant_id"`
	ProjectID     string                `json:"project_id"`
	EnvironmentID string                `json:"environment_id"`
	ParentGrantID string                `json:"parent_grant_id"`
	RequestedTool string                `json:"requested_tool"`
	Depth         int                   `json:"depth"`
	Budget        AISupervisorBudget    `json:"budget"`
	ContextHash   string                `json:"context_hash"`
}

type AISupervisorSubResult struct {
	SubgoalID   string                `json:"subgoal_id"`
	AgentRole   AISupervisorAgentRole `json:"agent_role"`
	Status      string                `json:"status"`
	SummaryCode string                `json:"summary_code"`
	Evidence    []AIEvidenceReference `json:"evidence"`
	PlanHash    string                `json:"plan_hash"`
	HighImpact  bool                  `json:"high_impact"`
}

type AISupervisorReport struct {
	SchemaVersion    string                  `json:"schema_version"`
	RunID            string                  `json:"run_id"`
	TenantID         string                  `json:"tenant_id"`
	ProjectID        string                  `json:"project_id"`
	ContextHash      string                  `json:"context_hash"`
	Subgoals         []AISupervisorSubgoal   `json:"subgoals"`
	Results          []AISupervisorSubResult `json:"results"`
	Conflicting      bool                    `json:"conflicting"`
	Escalated        bool                    `json:"escalated"`
	EscalationReason string                  `json:"escalation_reason,omitempty"`
	CostMicros       int64                   `json:"cost_micros"`
	GeneratedAt      time.Time               `json:"generated_at"`
}

func (b AISupervisorBudget) Validate() error {
	if b.MaxDepth < 0 || b.MaxDepth > 3 || b.MaxFanOut < 1 || b.MaxFanOut > 8 || b.MaxParallel < 1 || b.MaxParallel > b.MaxFanOut || b.MaxCostMicros < 0 || b.MaxWallTimeSeconds < 1 || b.MaxWallTimeSeconds > 3600 {
		return errors.New("supervisor budget is invalid")
	}
	return nil
}

func (s AISupervisorSubgoal) Validate(parent *AIAgentGrant) error {
	if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.GoalCode) == "" || strings.TrimSpace(s.TenantID) == "" || strings.TrimSpace(s.ProjectID) == "" || strings.TrimSpace(s.EnvironmentID) == "" || strings.TrimSpace(s.ParentGrantID) == "" || strings.TrimSpace(s.RequestedTool) == "" || strings.TrimSpace(s.ContextHash) == "" || s.Depth < 0 {
		return errors.New("supervisor subgoal identity is invalid")
	}
	switch s.Role {
	case AISupervisorDiagnosis, AISupervisorSecurity, AISupervisorGitOps, AISupervisorKubernetes, AISupervisorFinOps, AISupervisorRelease:
	default:
		return errors.New("unregistered supervisor role")
	}
	if err := s.Budget.Validate(); err != nil {
		return err
	}
	if parent != nil {
		if s.ParentGrantID != parent.ID || s.TenantID != parent.TenantID || s.ProjectID != parent.ProjectID || s.EnvironmentID != parent.ResourceID || s.RequestedTool != parent.ToolName || s.Depth != 1 || !s.ContextHashMatch(parent.ContextHash) {
			return errors.New("subgoal exceeds parent grant")
		}
	}
	return nil
}

func (s AISupervisorSubgoal) ContextHashMatch(parent string) bool {
	return strings.TrimSpace(parent) != "" && s.ContextHash == parent
}

func (r AISupervisorSubResult) Validate(tenantID string) error {
	if strings.TrimSpace(r.SubgoalID) == "" || strings.TrimSpace(r.SummaryCode) == "" || strings.TrimSpace(r.PlanHash) == "" || len(r.Evidence) == 0 {
		return errors.New("supervisor result provenance is incomplete")
	}
	for _, e := range r.Evidence {
		if e.TenantID != tenantID || e.Validate() != nil {
			return errors.New("supervisor result evidence is outside tenant scope")
		}
	}
	return nil
}

func (r AISupervisorReport) Validate() error {
	if r.SchemaVersion != AISupervisorSchemaVersion || strings.TrimSpace(r.RunID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.ContextHash) == "" || len(r.Subgoals) == 0 || len(r.Results) != len(r.Subgoals) || r.CostMicros < 0 || r.GeneratedAt.IsZero() {
		return errors.New("supervisor report identity is invalid")
	}
	seen := map[string]bool{}
	for _, subgoal := range r.Subgoals {
		if seen[subgoal.ID] {
			return errors.New("duplicate supervisor subgoal")
		}
		seen[subgoal.ID] = true
		if subgoal.TenantID != r.TenantID || subgoal.ProjectID != r.ProjectID {
			return errors.New("supervisor subgoal escapes scope")
		}
	}
	for _, result := range r.Results {
		if err := result.Validate(r.TenantID); err != nil {
			return err
		}
		if !seen[result.SubgoalID] {
			return errors.New("supervisor result has unknown subgoal")
		}
	}
	if r.Conflicting && !r.Escalated {
		return errors.New("conflicting high-impact results require escalation")
	}
	if r.Escalated && strings.TrimSpace(r.EscalationReason) == "" {
		return errors.New("supervisor escalation reason is required")
	}
	return nil
}
