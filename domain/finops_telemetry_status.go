package domain

import "time"

type FinOpsDimensionSummary struct {
	UsedBytes         *int64          `json:"usedBytes,omitempty"`
	Dimension         FinOpsDimension `json:"dimension"`
	State             string          `json:"state"`
	Reason            string          `json:"reason,omitempty"`
	Unit              string          `json:"unit"`
	MeasurementKind   string          `json:"measurementKind"`
	Source            string          `json:"source"`
	ExpectedResources int             `json:"expectedResources"`
	ObservedResources int             `json:"observedResources"`
	Quantity          float64         `json:"quantity"`
}

type FinOpsTelemetryWindow struct {
	ProjectID        string                   `json:"projectId"`
	ClusterID        string                   `json:"clusterId"`
	PeriodStart      time.Time                `json:"periodStart"`
	PeriodEnd        time.Time                `json:"periodEnd"`
	Fresh            bool                     `json:"fresh"`
	MetricsAvailable bool                     `json:"metricsAvailable"`
	ExpectedPods     int                      `json:"expectedPods"`
	MeasuredPods     int                      `json:"measuredPods"`
	UnattributedPods int                      `json:"unattributedPods"`
	CPUCoreHours     float64                  `json:"cpuCoreHours"`
	MemoryGiBHours   float64                  `json:"memoryGiBHours"`
	Dimensions       []FinOpsDimensionSummary `json:"dimensions"`
}

// Measured usage is not a provider invoice or a fabricated total spend.
type FinOpsTelemetryStatus struct {
	ObservedAt            time.Time               `json:"observedAt"`
	Scope                 string                  `json:"scope"`
	TotalCostKnown        bool                    `json:"totalCostKnown"`
	CostUnavailableReason string                  `json:"costUnavailableReason"`
	Windows               []FinOpsTelemetryWindow `json:"windows"`
}
