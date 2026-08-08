package domain

import (
	"encoding/json"
	"time"
)

// AuditEntry is the stable, storage-independent representation of a control
// plane action. Payload must contain metadata only and must never contain raw
// credentials.
type AuditEntry struct {
	ID         int64           `json:"id,omitempty"`
	OccurredAt time.Time       `json:"occurredAt"`
	Actor      string          `json:"actor"`
	Action     string          `json:"action"`
	Resource   string          `json:"resource"`
	ResourceID string          `json:"resourceId,omitempty"`
	ProjectID  string          `json:"projectId,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

// AuditFilter defines deterministic, newest-first pagination for audit reads.
type AuditFilter struct {
	Action     string
	Resource   string
	ResourceID string
	ProjectID  string
	Limit      int
	Offset     int
}
