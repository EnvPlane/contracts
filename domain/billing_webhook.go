package domain

import "time"

type BillingWebhookEvent struct {
	Provider    string    `json:"provider"`
	EventID     string    `json:"eventId"`
	TenantID    string    `json:"tenantId"`
	EventType   string    `json:"eventType"`
	PayloadHash string    `json:"payloadHash"`
	PayloadRef  string    `json:"payloadRef"`
	ReceivedAt  time.Time `json:"receivedAt"`
	Status      string    `json:"status"`
	Attempts    int       `json:"attempts"`
	LastError   string    `json:"lastError,omitempty"`
}
