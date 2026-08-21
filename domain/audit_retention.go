package domain

import "time"

type AuditRetentionPolicy struct {
	TenantID      string `json:"tenantId"`
	RetentionDays int64  `json:"retentionDays"`
	LegalHold     bool   `json:"legalHold"`
}

type AuditRetentionPlan struct {
	TenantID    string    `json:"tenantId"`
	PurgeBefore time.Time `json:"purgeBefore"`
	Candidates  []int64   `json:"candidates"`
	Bounded     bool      `json:"bounded"`
	LegalHold   bool      `json:"legalHold"`
}
