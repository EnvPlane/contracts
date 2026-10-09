package domain

import "time"

// TenantFinOpsPriceVersion is an immutable operator-supplied chargeback table.
// Nil rates mean unknown; an explicitly supplied zero is a known zero price.
// It is not a tax determination, provider invoice or subscription offer.
type TenantFinOpsPriceVersion struct {
	Revision           int64                  `json:"revision"`
	Currency           string                 `json:"currency"`
	Source             string                 `json:"source"`
	EffectiveFrom      time.Time              `json:"effectiveFrom"`
	CreatedAt          time.Time              `json:"createdAt"`
	CPUCoreHourCents   *int64                 `json:"cpuCoreHourCents"`
	MemoryGiBHourCents *int64                 `json:"memoryGiBHourCents"`
	CPUCoreHourRate    *FinOpsScaledRate      `json:"cpuCoreHourRate,omitempty"`
	MemoryGiBHourRate  *FinOpsScaledRate      `json:"memoryGiBHourRate,omitempty"`
	Dimensions         []FinOpsDimensionPrice `json:"dimensions"`
}
