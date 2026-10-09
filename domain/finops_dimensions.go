package domain

import "time"

type FinOpsDimension string

const (
	FinOpsStorageRequested   FinOpsDimension = "storage.requested"
	FinOpsStorageProvisioned FinOpsDimension = "storage.provisioned"
	FinOpsStorageUsed        FinOpsDimension = "storage.used"
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
	UsedBytes     *int64  `json:"usedBytes,omitempty"`
	SampleID      string  `json:"sampleId"`
	TenantID      string  `json:"tenantId,omitempty"`
	EnvironmentID string  `json:"environmentId"`
	ComponentID   string  `json:"componentId"`
	Namespace     string  `json:"namespace"`
	ResourceUID   string  `json:"resourceUid"`
	Quantity      float64 `json:"quantity"`
}

type FinOpsDimensionReport struct {
	GPUInventory      *FinOpsGPUInventory     `json:"gpuInventory,omitempty"`
	Dimension         FinOpsDimension         `json:"dimension"`
	Unit              string                  `json:"unit"`
	MeasurementKind   string                  `json:"measurementKind"`
	Source            string                  `json:"source"`
	PeriodStart       time.Time               `json:"periodStart"`
	PeriodEnd         time.Time               `json:"periodEnd"`
	ExpectedResources int                     `json:"expectedResources"`
	ObservedResources int                     `json:"observedResources"`
	State             string                  `json:"state"` // complete, partial, unavailable, not_applicable
	Reason            string                  `json:"reason,omitempty"`
	Samples           []FinOpsDimensionSample `json:"samples"`
}

// GPUInventory must come from an authenticated, complete cluster-node inventory,
// not absence of exporter series. Receiver verifies cluster binding/freshness.
type FinOpsGPUInventory struct {
	ClusterID  string    `json:"clusterId"`
	ObservedAt time.Time `json:"observedAt"`
	Nodes      int       `json:"nodes"`
	Devices    int       `json:"devices"`
	Source     string    `json:"source"`
}

type FinOpsDimensionPrice struct {
	Dimension         FinOpsDimension `json:"dimension"`
	Unit              string          `json:"unit"`
	Currency          string          `json:"currency"`
	MinorUnitsPerUnit int64           `json:"minorUnitsPerUnit"`
	Known             bool            `json:"known"`
	// ScaledRate is exclusive with legacy Known=true. Old readers see UNKNOWN,
	// never a fabricated zero or an unscaled numerator as a whole-cent price.
	ScaledRate *FinOpsScaledRate `json:"scaledRate,omitempty"`
}

type FinOpsDimensionAllocation struct {
	PriceRevision int64  `json:"priceRevision,omitempty"`
	PriceSource   string `json:"priceSource,omitempty"`
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

// Resource estimates do not cover idle nodes, managed/shared services or a
// reconciled cloud invoice. Missing sources/prices must be visible separately.
type FinOpsCostEvidence struct {
	Scope                        string            `json:"scope"`
	Partial                      bool              `json:"partial"`
	TotalInfrastructureCostKnown bool              `json:"totalInfrastructureCostKnown"`
	UnavailableDimensions        []FinOpsDimension `json:"unavailableDimensions"`
	UnpricedDimensions           []FinOpsDimension `json:"unpricedDimensions"`
}
