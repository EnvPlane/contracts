package domain

import "time"

type MonthlyBudget struct {
	TenantID    string    `json:"tenantId"`
	Currency    string    `json:"currency"`
	AmountCents int64     `json:"amountCents"`
	MonthStart  time.Time `json:"monthStart"`
}
type BudgetForecast struct {
	TenantID      string    `json:"tenantId"`
	Currency      string    `json:"currency"`
	ActualCents   int64     `json:"actualCents"`
	ForecastCents int64     `json:"forecastCents"`
	PeriodStart   time.Time `json:"periodStart"`
	PeriodEnd     time.Time `json:"periodEnd"`
	Method        string    `json:"method"`
}
type BudgetAnomaly struct {
	TenantID       string    `json:"tenantId"`
	RuleID         string    `json:"ruleId"`
	Severity       string    `json:"severity"`
	DeviationCents int64     `json:"deviationCents"`
	Message        string    `json:"message"`
	ObservedAt     time.Time `json:"observedAt"`
}
