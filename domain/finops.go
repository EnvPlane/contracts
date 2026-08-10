package domain

import "time"

type ResourceUsageSample struct {
	SnapshotID     string    `json:"snapshotId"`
	TenantID       string    `json:"tenantId"`
	ProjectID      string    `json:"projectId,omitempty"`
	EnvironmentID  string    `json:"environmentId,omitempty"`
	ComponentID    string    `json:"componentId,omitempty"`
	CPUCoreHours   float64   `json:"cpuCoreHours"`
	MemoryGiBHours float64   `json:"memoryGiBHours"`
	OccurredAt     time.Time `json:"occurredAt"`
}

type ResourcePriceTable struct {
	Currency           string `json:"currency"`
	CPUCoreHourCents   int64  `json:"cpuCoreHourCents"`
	MemoryGiBHourCents int64  `json:"memoryGiBHourCents"`
}

type CostAllocation struct {
	SnapshotID     string    `json:"snapshotId"`
	TenantID       string    `json:"tenantId"`
	ProjectID      string    `json:"projectId,omitempty"`
	EnvironmentID  string    `json:"environmentId,omitempty"`
	ComponentID    string    `json:"componentId,omitempty"`
	Currency       string    `json:"currency"`
	CPUCoreHours   float64   `json:"cpuCoreHours"`
	MemoryGiBHours float64   `json:"memoryGiBHours"`
	AmountCents    int64     `json:"amountCents"`
	PriceKnown     bool      `json:"priceKnown"`
	PeriodStart    time.Time `json:"periodStart"`
	PeriodEnd      time.Time `json:"periodEnd"`
}
