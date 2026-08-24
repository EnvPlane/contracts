package domain

import (
	"errors"
	"strings"
	"time"
)

const AITriggerSchemaVersion = "1"

type AITriggerRule struct {
	SchemaVersion   string          `json:"schema_version"`
	ID              string          `json:"id"`
	TenantID        string          `json:"tenant_id"`
	Purpose         string          `json:"purpose"`
	EventTypes      []string        `json:"event_types"`
	FilterCodes     []string        `json:"filter_codes,omitempty"`
	CooldownSeconds int             `json:"cooldown_seconds"`
	Timezone        string          `json:"timezone"`
	QuietStart      string          `json:"quiet_start,omitempty"`
	QuietEnd        string          `json:"quiet_end,omitempty"`
	MaintenanceOnly bool            `json:"maintenance_only"`
	Enabled         bool            `json:"enabled"`
	Autonomy        AIAutonomyLevel `json:"autonomy"`
}

type AITriggerEvent struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Type        string    `json:"type"`
	FilterCode  string    `json:"filter_code,omitempty"`
	SubjectID   string    `json:"subject_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Fingerprint string    `json:"fingerprint"`
}

type AITriggerState struct {
	LastEventID     string    `json:"last_event_id,omitempty"`
	LastFingerprint string    `json:"last_fingerprint,omitempty"`
	LastTriggeredAt time.Time `json:"last_triggered_at,omitempty"`
	StormCount      int       `json:"storm_count"`
}

type AITriggerDecision string

const (
	AITriggerDisabled     AITriggerDecision = "disabled"
	AITriggerDeduplicated AITriggerDecision = "deduplicated"
	AITriggerCooldown     AITriggerDecision = "cooldown"
	AITriggerQuietHours   AITriggerDecision = "quiet_hours"
	AITriggerPreview      AITriggerDecision = "preview"
	AITriggerStarted      AITriggerDecision = "started"
	AITriggerEscalated    AITriggerDecision = "escalated"
)

type AITriggerResult struct {
	SchemaVersion string            `json:"schema_version"`
	RuleID        string            `json:"rule_id"`
	EventID       string            `json:"event_id"`
	TenantID      string            `json:"tenant_id"`
	Decision      AITriggerDecision `json:"decision"`
	RunKey        string            `json:"run_key"`
	Reason        string            `json:"reason"`
	Autonomy      AIAutonomyLevel   `json:"autonomy"`
	ScheduledAt   *time.Time        `json:"scheduled_at,omitempty"`
	State         AITriggerState    `json:"state"`
}

type AITriggerPreviewRequest struct {
	Rule  AITriggerRule  `json:"rule"`
	Event AITriggerEvent `json:"event"`
	State AITriggerState `json:"state"`
}

func (r AITriggerPreviewRequest) Validate(tenantID string) error {
	if r.Rule.TenantID != tenantID {
		return errors.New("trigger rule is outside tenant scope")
	}
	return r.Rule.Validate()
}

func (r AITriggerRule) Validate() error {
	if r.SchemaVersion != AITriggerSchemaVersion || strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.Purpose) == "" || len(r.EventTypes) == 0 || len(r.EventTypes) > 32 || r.CooldownSeconds < 0 || r.CooldownSeconds > 7*24*3600 || strings.TrimSpace(r.Timezone) == "" {
		return errors.New("trigger rule identity or bounds are invalid")
	}
	if r.Autonomy != AIAutonomyObserve && r.Autonomy != AIAutonomyRecommend && r.Autonomy != AIAutonomyApprovalRequired && r.Autonomy != AIAutonomyAutonomousBounded {
		return errors.New("trigger autonomy is invalid")
	}
	return nil
}

func (e AITriggerEvent) Validate(tenantID string) error {
	if strings.TrimSpace(e.ID) == "" || e.TenantID != tenantID || strings.TrimSpace(e.Type) == "" || strings.TrimSpace(e.SubjectID) == "" || strings.TrimSpace(e.Fingerprint) == "" || e.OccurredAt.IsZero() {
		return errors.New("trigger event is invalid or outside tenant scope")
	}
	return nil
}
