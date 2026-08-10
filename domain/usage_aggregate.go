package domain

import "time"

type UsageAggregate struct {
	TenantID    string    `json:"tenantId"`
	Metric      string    `json:"metric"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	Quantity    int64     `json:"quantity"`
	EventCount  int64     `json:"eventCount"`
	Revision    string    `json:"revision"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UsageCheckpoint struct {
	TenantID     string    `json:"tenantId"`
	Cursor       time.Time `json:"cursor"`
	LastEventID  string    `json:"lastEventId"`
	ReconciledAt time.Time `json:"reconciledAt"`
}
