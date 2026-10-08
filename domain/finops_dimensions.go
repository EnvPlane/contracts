package domain

import "time"

type FinOpsDimension string

const (
	FinOpsStorageRequested   FinOpsDimension = "storage.requested"
	FinOpsStorageProvisioned FinOpsDimension = "storage.provisioned"
	FinOpsNetworkTransmit    FinOpsDimension = "network.transmit"
	FinOpsNetworkReceive     FinOpsDimension = "network.receive"
	FinOpsGPUUtilization     FinOpsDimension = "gpu.utilization"
	FinOpsGiBHours                           = "gib_hours"
	FinOpsGiB                                = "gib"
	FinOpsBusyGPUHours                       = "busy_gpu_hours"
)

// Each dimension has independent provenance, coverage and price semantics.
// Provisioned storage is capacity, not bytes used or a cloud-provider invoice.
type FinOpsDimensionSample struct {
	SampleID      string  `json:"sampleId"`
	TenantID      string  `json:"tenantId,omitempty"`
	EnvironmentID string  `json:"environmentId"`
	ComponentID   string  `json:"componentId"`
	Namespace     string  `json:"namespace"`
	ResourceUID   string  `json:"resourceUid"`
	Quantity      float64 `json:"quantity"`
}

type FinOpsDimensionReport struct {
	Dimension         FinOpsDimension         `json:"dimension"`
	Unit              string                  `json:"unit"`
	MeasurementKind   string                  `json:"measurementKind"`
	Source            string                  `json:"source"`
	PeriodStart       time.Time               `json:"periodStart"`
	PeriodEnd         time.Time               `json:"periodEnd"`
	ExpectedResources int                     `json:"expectedResources"`
	ObservedResources int                     `json:"observedResources"`
	State             string                  `json:"state"` // complete, partial, unavailable; never inferred from sample count
	Reason            string                  `json:"reason,omitempty"`
	Samples           []FinOpsDimensionSample `json:"samples"`
}

type FinOpsDimensionPrice struct {
	Dimension         FinOpsDimension `json:"dimension"`
	Unit              string          `json:"unit"`
	Currency          string          `json:"currency"`
	MinorUnitsPerUnit int64           `json:"minorUnitsPerUnit"`
	Known             bool            `json:"known"`
}

type FinOpsDimensionAllocation struct {
	FinOpsDimensionSample
	Dimension       FinOpsDimension `json:"dimension"`
	Unit            string          `json:"unit"`
	MeasurementKind string          `json:"measurementKind"`
	Source          string          `json:"source"`
	ProjectID       string          `json:"projectId"`
	ClusterID       string          `json:"clusterId"`
	PeriodStart     time.Time       `json:"periodStart"`
	PeriodEnd       time.Time       `json:"periodEnd"`
	Cost            Money           `json:"cost"`
}
