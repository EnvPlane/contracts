package domain

import "time"

// Baseline evidence is a separate non-financial ledger, not Environment spend.
type BaselineMetricCoverage struct {
	BindingID       string  `json:"bindingId"`
	BindingVersion  int64   `json:"bindingVersion"`
	ResourceUID     string  `json:"resourceUid"`
	Metric          string  `json:"metric"`
	State           string  `json:"state"`
	Reason          string  `json:"reason,omitempty"`
	CoveredSeconds  float64 `json:"coveredSeconds"`
	ExpectedSeconds float64 `json:"expectedSeconds"`
	Samples         int     `json:"samples"`
}
type BaselineMeteringEvidence struct {
	ProjectID         string                   `json:"projectId"`
	ClusterID         string                   `json:"clusterId"`
	ClusterGeneration int64                    `json:"clusterGeneration"`
	PeriodStart       time.Time                `json:"periodStart"`
	PeriodEnd         time.Time                `json:"periodEnd"`
	Receipts          []BaselineMeteringBatch  `json:"receipts"`
	Coverage          []BaselineMetricCoverage `json:"coverage"`
	Truncated         bool                     `json:"truncated"`
	CostKnown         bool                     `json:"costKnown"`
	BudgetIncluded    bool                     `json:"budgetIncluded"`
}
