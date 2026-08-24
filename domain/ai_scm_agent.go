package domain

import (
	"errors"
	"strings"
	"time"
)

const AISCMChangePlanSchemaVersion = "1"

type AISCMImpact string

const (
	AISCMImpactUnknown AISCMImpact = "unknown"
	AISCMImpactCreate  AISCMImpact = "environment_create_or_update"
	AISCMImpactCleanup AISCMImpact = "policy_controlled_cleanup"
	AISCMImpactIgnored AISCMImpact = "ignored"
)

type AISCMChangePlan struct {
	SchemaVersion    string      `json:"schemaVersion"`
	PlanID           string      `json:"planId"`
	TenantID         string      `json:"tenantId"`
	ProjectID        string      `json:"projectId"`
	Provider         Provider    `json:"provider"`
	Repository       string      `json:"repository"`
	ChangeID         string      `json:"changeId"`
	EventID          string      `json:"eventId"`
	Action           EventAction `json:"action"`
	Branch           string      `json:"branch,omitempty"`
	CommitSHA        string      `json:"commitSha,omitempty"`
	ChangedPaths     []string    `json:"changedPaths,omitempty"`
	Impact           AISCMImpact `json:"impact"`
	UntrustedInput   bool        `json:"untrustedInput"`
	ApprovalRequired bool        `json:"approvalRequired"`
	CleanupAllowed   bool        `json:"cleanupAllowed"`
	IdempotencyKey   string      `json:"idempotencyKey"`
	GeneratedAt      time.Time   `json:"generatedAt"`
}

func (p AISCMChangePlan) Validate() error {
	if p.SchemaVersion != AISCMChangePlanSchemaVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.Repository) == "" || strings.TrimSpace(p.ChangeID) == "" || strings.TrimSpace(p.EventID) == "" || strings.TrimSpace(p.IdempotencyKey) == "" || p.GeneratedAt.IsZero() {
		return errors.New("SCM change plan identity is invalid")
	}
	if p.Provider != ProviderGitHub && p.Provider != ProviderGitLab {
		return errors.New("unsupported SCM provider")
	}
	if p.Action != ActionOpen && p.Action != ActionUpdate && p.Action != ActionClose && p.Action != ActionIgnore {
		return errors.New("unsupported SCM action")
	}
	for _, value := range []string{p.Repository, p.ChangeID, p.EventID, p.Branch, p.CommitSHA, p.PlanID, p.IdempotencyKey} {
		if len(value) > 1024 || strings.ContainsAny(value, "\r\n\x00") {
			return errors.New("SCM plan contains invalid untrusted metadata")
		}
	}
	if len(p.ChangedPaths) > 256 {
		return errors.New("SCM changed path limit exceeded")
	}
	for _, path := range p.ChangedPaths {
		if len(path) > 512 || strings.ContainsAny(path, "\r\n\x00") || strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
			return errors.New("SCM changed path is invalid")
		}
	}
	return nil
}
