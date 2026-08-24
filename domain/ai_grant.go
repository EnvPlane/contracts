package domain

import (
	"errors"
	"strings"
	"time"
)

const AIAgentGrantSchemaVersion = "1"

type AIGrantRevokeReason string

const (
	AIGrantRevokedCancellation AIGrantRevokeReason = "run_canceled"
	AIGrantRevokedPolicy       AIGrantRevokeReason = "policy_changed"
	AIGrantRevokedApproval     AIGrantRevokeReason = "approval_expired"
	AIGrantRevokedManual       AIGrantRevokeReason = "manual_revoke"
)

// AIAgentGrant is an opaque, non-exportable capability record. The grant ID
// is only a lookup handle; it is not a credential and carries no secret.
type AIAgentGrant struct {
	SchemaVersion string             `json:"schemaVersion"`
	ID            string             `json:"id"`
	RunID         string             `json:"runId"`
	TenantID      string             `json:"tenantId"`
	ProjectID     string             `json:"projectId"`
	ResourceType  string             `json:"resourceType"`
	ResourceID    string             `json:"resourceId"`
	ToolName      string             `json:"toolName"`
	Purpose       string             `json:"purpose"`
	RequesterID   string             `json:"requesterId"`
	ApproverID    string             `json:"approverId,omitempty"`
	AgentID       string             `json:"agentId"`
	ContextHash   string             `json:"contextHash"`
	IssuedAt      time.Time          `json:"issuedAt"`
	ExpiresAt     time.Time          `json:"expiresAt"`
	RevokedAt     *time.Time         `json:"revokedAt,omitempty"`
	RevokeReason  AIGrantRevokeReason `json:"revokeReason,omitempty"`
}

func (g AIAgentGrant) Validate(now time.Time) error {
	if g.SchemaVersion != AIAgentGrantSchemaVersion || strings.TrimSpace(g.ID) == "" || strings.TrimSpace(g.RunID) == "" || strings.TrimSpace(g.TenantID) == "" || strings.TrimSpace(g.ProjectID) == "" || strings.TrimSpace(g.ResourceType) == "" || strings.TrimSpace(g.ResourceID) == "" || strings.TrimSpace(g.ToolName) == "" || strings.TrimSpace(g.Purpose) == "" || strings.TrimSpace(g.RequesterID) == "" || strings.TrimSpace(g.AgentID) == "" || strings.TrimSpace(g.ContextHash) == "" {
		return errors.New("AI agent grant identity and scope are required")
	}
	if g.IssuedAt.IsZero() || g.ExpiresAt.IsZero() || !g.ExpiresAt.After(g.IssuedAt) {
		return errors.New("AI agent grant lifetime is invalid")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if g.ExpiresAt.Before(now.UTC()) {
		return errors.New("AI agent grant is expired")
	}
	if g.RevokedAt != nil {
		if g.RevokeReason == "" {
			return errors.New("AI agent grant revoke reason is required")
		}
		return errors.New("AI agent grant is revoked")
	}
	return nil
}
