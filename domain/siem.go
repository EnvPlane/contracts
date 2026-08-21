package domain

import "time"

type SIEMAuditEvent struct {
	EventID    string    `json:"eventId"`
	TenantID   string    `json:"tenantId"`
	Cursor     int64     `json:"cursor,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
	Actor      string    `json:"actor"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resourceId,omitempty"`
}

type SIEMExportCheckpoint struct {
	TenantID string `json:"tenantId"`
	Cursor   int64  `json:"cursor"`
}
