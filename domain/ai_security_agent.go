package domain

import (
	"errors"
	"strings"
	"time"
)

const AISecurityAgentSchemaVersion = "1"

type AISecuritySeverity string

const (
	AISecurityInfo     AISecuritySeverity = "info"
	AISecurityLow      AISecuritySeverity = "low"
	AISecurityMedium   AISecuritySeverity = "medium"
	AISecurityHigh     AISecuritySeverity = "high"
	AISecurityCritical AISecuritySeverity = "critical"
)

type AISecurityFindingKind string

const (
	AISecurityRBACEscalation     AISecurityFindingKind = "rbac_escalation"
	AISecurityUnsignedImage      AISecurityFindingKind = "unsigned_image"
	AISecurityVulnerableArtifact AISecurityFindingKind = "vulnerable_artifact"
	AISecurityNetworkPolicyDrift AISecurityFindingKind = "network_policy_drift"
	AISecurityPolicyDrift        AISecurityFindingKind = "policy_drift"
	AISecurityAuditAnomaly       AISecurityFindingKind = "audit_anomaly"
)

type AISecurityException struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	ProjectID  string    `json:"project_id"`
	FindingID  string    `json:"finding_id"`
	Reason     string    `json:"reason"`
	ApproverID string    `json:"approver_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	Audited    bool      `json:"audited"`
}

type AISecurityFinding struct {
	ID            string                `json:"id"`
	Kind          AISecurityFindingKind `json:"kind"`
	Severity      AISecuritySeverity    `json:"severity"`
	TenantID      string                `json:"tenant_id"`
	ProjectID     string                `json:"project_id"`
	EnvironmentID string                `json:"environment_id"`
	Source        AIEvidenceReference   `json:"source"`
	Subject       string                `json:"subject"`
	Digest        string                `json:"digest,omitempty"`
	Exception     *AISecurityException  `json:"exception,omitempty"`
}

type AISecurityActionKind string

const (
	AISecurityPolicyDiff        AISecurityActionKind = "policy_diff"
	AISecurityRotateCredential  AISecurityActionKind = "rotate_credential" //nolint:gosec -- action identifier, not credential material
	AISecurityRBACMutation      AISecurityActionKind = "rbac_mutation"
	AISecurityNetworkPolicyDiff AISecurityActionKind = "network_policy_diff"
)

type AISecurityRemediationProposal struct {
	Kind             AISecurityActionKind `json:"kind"`
	FindingID        string               `json:"finding_id"`
	Target           string               `json:"target"`
	Diff             string               `json:"diff"`
	ApprovalRequired bool                 `json:"approval_required"`
	Preconditions    []string             `json:"preconditions"`
	Verification     []string             `json:"verification"`
	Compensation     string               `json:"compensation"`
}

type AISecurityAgentRequest struct {
	ProjectID     string              `json:"project_id"`
	EnvironmentID string              `json:"environment_id"`
	Findings      []AISecurityFinding `json:"findings"`
}

type AISecurityAgentPlan struct {
	SchemaVersion    string                          `json:"schema_version"`
	PlanID           string                          `json:"plan_id"`
	TenantID         string                          `json:"tenant_id"`
	ProjectID        string                          `json:"project_id"`
	EnvironmentID    string                          `json:"environment_id"`
	ContextHash      string                          `json:"context_hash"`
	Findings         []AISecurityFinding             `json:"findings"`
	Evidence         []AIEvidenceReference           `json:"evidence"`
	Proposals        []AISecurityRemediationProposal `json:"proposals"`
	FailClosed       bool                            `json:"fail_closed"`
	EscalationReason string                          `json:"escalation_reason,omitempty"`
	GeneratedAt      time.Time                       `json:"generated_at"`
}

func securityStringSafe(value string) bool {
	value = strings.ToLower(value)
	for _, marker := range []string{"secret", "password", "token", "kubeconfig", "private key", "authorization:"} {
		if strings.Contains(value, marker) {
			return false
		}
	}
	return !strings.ContainsAny(value, "\r\n\x00")
}

func (e AISecurityException) Validate(tenantID, projectID, findingID string, now time.Time) error {
	if !e.Audited || e.TenantID != tenantID || e.ProjectID != projectID || e.FindingID != findingID || strings.TrimSpace(e.ID) == "" || strings.TrimSpace(e.Reason) == "" || strings.TrimSpace(e.ApproverID) == "" || e.ExpiresAt.Before(now) {
		return errors.New("security exception is missing audited ownership or has expired")
	}
	return nil
}

func (f AISecurityFinding) Validate(tenantID, projectID, environmentID string, now time.Time) error {
	if strings.TrimSpace(f.ID) == "" || f.TenantID != tenantID || f.ProjectID != projectID || f.EnvironmentID != environmentID || f.Source.TenantID != tenantID || f.Source.SourceType == "" || f.Source.SourceID == "" || strings.TrimSpace(f.Subject) == "" || !securityStringSafe(f.Subject) || !securityStringSafe(f.Digest) {
		return errors.New("security finding is outside scope or contains unsafe data")
	}
	if f.Severity != AISecurityInfo && f.Severity != AISecurityLow && f.Severity != AISecurityMedium && f.Severity != AISecurityHigh && f.Severity != AISecurityCritical {
		return errors.New("unknown security severity")
	}
	if f.Exception != nil {
		if err := f.Exception.Validate(tenantID, projectID, f.ID, now); err != nil {
			return err
		}
	}
	return nil
}

func (r AISecurityAgentRequest) Validate(tenantID string, now time.Time) error {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.EnvironmentID) == "" || len(r.Findings) == 0 || len(r.Findings) > 256 {
		return errors.New("security request scope or bounds are invalid")
	}
	for _, finding := range r.Findings {
		if err := finding.Validate(tenantID, r.ProjectID, r.EnvironmentID, now); err != nil {
			return err
		}
	}
	return nil
}

func (p AISecurityAgentPlan) Validate(now time.Time) error {
	if p.SchemaVersion != AISecurityAgentSchemaVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.EnvironmentID) == "" || strings.TrimSpace(p.ContextHash) == "" || len(p.Findings) == 0 || len(p.Evidence) == 0 || p.GeneratedAt.IsZero() {
		return errors.New("security plan identity is invalid")
	}
	for _, finding := range p.Findings {
		if err := finding.Validate(p.TenantID, p.ProjectID, p.EnvironmentID, now); err != nil {
			return err
		}
	}
	for _, evidence := range p.Evidence {
		if evidence.TenantID != p.TenantID || evidence.Validate() != nil {
			return errors.New("security evidence is outside tenant scope")
		}
	}
	if p.FailClosed && strings.TrimSpace(p.EscalationReason) == "" {
		return errors.New("fail-closed security plan needs escalation reason")
	}
	for _, proposal := range p.Proposals {
		if strings.TrimSpace(proposal.FindingID) == "" || strings.TrimSpace(proposal.Target) == "" || len(proposal.Preconditions) == 0 || len(proposal.Verification) == 0 || strings.TrimSpace(proposal.Compensation) == "" || !securityStringSafe(proposal.Diff) {
			return errors.New("security proposal lacks bounded remediation")
		}
		if proposal.Kind == AISecurityRotateCredential || proposal.Kind == AISecurityRBACMutation || !proposal.ApprovalRequired {
			if proposal.Kind == AISecurityRotateCredential || proposal.Kind == AISecurityRBACMutation {
				if !proposal.ApprovalRequired {
					return errors.New("security mutation requires approval")
				}
			}
		}
	}
	return nil
}
