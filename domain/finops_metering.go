package domain

import "time"

const (
	FinOpsMeasured = "measured"
	FinOpsCapacity = "capacity"
)

// FinOpsMeteringBatch carries no authoritative tenant identity. The receiving
// API derives tenant from the current authenticated Agent/project binding.
type FinOpsMeteringBatch struct {
	Dimensions       []FinOpsDimensionReport `json:"dimensions,omitempty"`
	BatchID          string                  `json:"batchId"`
	ProjectID        string                  `json:"projectId"`
	ClusterID        string                  `json:"clusterId"`
	AgentID          string                  `json:"agentId"`
	PeriodStart      time.Time               `json:"periodStart"`
	PeriodEnd        time.Time               `json:"periodEnd"`
	ExpectedPods     int                     `json:"expectedPods"`
	MeasuredPods     int                     `json:"measuredPods"`
	UnattributedPods int                     `json:"unattributedPods"`
	MetricsAvailable bool                    `json:"metricsAvailable"`
	Samples          []ResourceUsageSample   `json:"samples"`
}

// Coverage is scoped to one authenticated project and cluster, not the entire
// tenant. Missing intervals or unregistered sources must never imply zero cost.
type FinOpsCoverage struct {
	ProjectID   string    `json:"projectId"`
	ClusterID   string    `json:"clusterId"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	Complete    bool      `json:"complete"`
	Reason      string    `json:"reason,omitempty"`
}
