package domain

import "time"

// Infrastructure reports contain exclusive reconciled monetary categories,
// not additive workload estimates over a node/provider bill. Operator evidence
// is explicit provenance, not a claim of automatic provider verification.
type InfrastructureCostLine struct {
	Category    string `json:"category"`
	State       string `json:"state"`
	AmountCents *int64 `json:"amountCents"`
}

type InfrastructureCostReport struct {
	TenantID    string                   `json:"tenantId"`
	Revision    int64                    `json:"revision"`
	Currency    string                   `json:"currency"`
	PeriodStart time.Time                `json:"periodStart"`
	PeriodEnd   time.Time                `json:"periodEnd"`
	ObservedAt  time.Time                `json:"observedAt"`
	Source      string                   `json:"source"`
	Reference   string                   `json:"reference"`
	Lines       []InfrastructureCostLine `json:"lines"`
}

var InfrastructureCostCategories = []string{"compute", "storage", "network", "gpu", "managed-services", "other"}
