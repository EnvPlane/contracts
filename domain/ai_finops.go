package domain

import (
	"errors"
	"strings"
	"time"
)

const AIFinOpsSchemaVersion = "1"

type FinOpsExplanationSnapshot struct {
	SchemaVersion   string
	TenantID        string
	ProjectID       string
	Currency        string
	Budget          MonthlyBudget
	Allocations     []CostAllocation
	PartialData     bool
	Recommendations []AIFinOpsRecommendation
}

type AIFinOpsRecommendation struct {
	ID             string                      `json:"id"`
	Reason         string                      `json:"reason"`
	ProposalFields ConfigurationProposalFields `json:"proposalFields"`
	Evidence       []AIEvidenceReference       `json:"evidence,omitempty"`
}

type AIFinOpsMoney struct {
	Currency   string `json:"currency"`
	MinorUnits int64  `json:"minorUnits"`
	Known      bool   `json:"known"`
}

type AIFinOpsBudgetVariance struct {
	Budget   AIFinOpsMoney `json:"budget"`
	Actual   AIFinOpsMoney `json:"actual"`
	Forecast AIFinOpsMoney `json:"forecast"`
	Variance AIFinOpsMoney `json:"variance"`
	Known    bool          `json:"known"`
}

type AIFinOpsForecast struct {
	Amount      AIFinOpsMoney `json:"amount"`
	PeriodStart time.Time     `json:"periodStart"`
	PeriodEnd   time.Time     `json:"periodEnd"`
	Method      string        `json:"method"`
	Formula     string        `json:"formula"`
	Known       bool          `json:"known"`
}

type AIFinOpsAnomalyEvidence struct {
	RuleID    string                `json:"ruleId"`
	Severity  string                `json:"severity"`
	Deviation AIFinOpsMoney         `json:"deviation"`
	Threshold int64                 `json:"thresholdPercent"`
	Evidence  []AIEvidenceReference `json:"evidence,omitempty"`
}

type AIFinOpsDataQuality struct {
	PriceKnown   bool `json:"priceKnown"`
	MissingPrice bool `json:"missingPrice"`
	PartialData  bool `json:"partialData"`
}

type AIFinOpsExplanation struct {
	SchemaVersion   string                    `json:"schemaVersion"`
	TenantID        string                    `json:"tenantId"`
	ProjectID       string                    `json:"projectId"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
	Currency        string                    `json:"currency"`
	PeriodStart     time.Time                 `json:"periodStart"`
	PeriodEnd       time.Time                 `json:"periodEnd"`
	Summary         string                    `json:"summary"`
	Variance        AIFinOpsBudgetVariance    `json:"variance"`
	Forecast        AIFinOpsForecast          `json:"forecast"`
	Anomalies       []AIFinOpsAnomalyEvidence `json:"anomalies,omitempty"`
	Recommendations []AIFinOpsRecommendation  `json:"recommendations,omitempty"`
	Evidence        []AIEvidenceReference     `json:"evidence,omitempty"`
	DataQuality     AIFinOpsDataQuality       `json:"dataQuality"`
}

func (e AIFinOpsExplanation) Validate() error {
	if e.SchemaVersion != AIFinOpsSchemaVersion || strings.TrimSpace(e.TenantID) == "" || strings.TrimSpace(e.ProjectID) == "" || strings.TrimSpace(e.Currency) == "" {
		return errors.New("invalid FinOps explanation identity or schema")
	}
	if e.PeriodStart.Location() != time.UTC || e.PeriodEnd.Location() != time.UTC || !e.PeriodEnd.After(e.PeriodStart) {
		return errors.New("FinOps explanation period must be ordered UTC timestamps")
	}
	if e.Variance.Known {
		if !e.Variance.Budget.Known || !e.Variance.Actual.Known || !e.Variance.Forecast.Known || !e.Variance.Variance.Known || e.Variance.Variance.MinorUnits != e.Variance.Forecast.MinorUnits-e.Variance.Budget.MinorUnits {
			return errors.New("FinOps variance is not grounded in structured values")
		}
	} else if e.Forecast.Known || e.Variance.Actual.Known {
		return errors.New("unknown FinOps variance cannot claim known values")
	}
	for _, evidence := range e.Evidence {
		if evidence.TenantID != e.TenantID || evidence.Validate() != nil {
			return errors.New("FinOps evidence is outside the explanation scope")
		}
	}
	for _, anomaly := range e.Anomalies {
		for _, evidence := range anomaly.Evidence {
			if evidence.TenantID != e.TenantID || evidence.Validate() != nil {
				return errors.New("FinOps anomaly evidence is outside the explanation scope")
			}
		}
	}
	for _, recommendation := range e.Recommendations {
		if strings.TrimSpace(recommendation.ID) == "" || strings.TrimSpace(recommendation.Reason) == "" || recommendation.ProposalFields.Validate() != nil {
			return errors.New("invalid typed FinOps recommendation")
		}
		for _, evidence := range recommendation.Evidence {
			if evidence.TenantID != e.TenantID || evidence.Validate() != nil {
				return errors.New("FinOps recommendation evidence is outside the explanation scope")
			}
		}
	}
	return nil
}
