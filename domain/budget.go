package domain

import "time"

const (
	BudgetEffectWarn = "warn"
	BudgetEffectDeny = "deny"
)

type MonthlyBudget struct {
	TenantID         string    `json:"tenantId"`
	Currency         string    `json:"currency"`
	AmountCents      int64     `json:"amountCents"`
	MonthStart       time.Time `json:"monthStart"`
	AnomalyThreshold int64     `json:"anomalyThresholdPercent"`
	Enforcement      string    `json:"enforcement"`
}
type BudgetForecast struct {
	TenantID      string    `json:"tenantId"`
	Currency      string    `json:"currency"`
	ActualCents   int64     `json:"actualCents"`
	ForecastCents int64     `json:"forecastCents"`
	BudgetCents   int64     `json:"budgetCents"`
	ElapsedHours  int64     `json:"elapsedHours"`
	PeriodHours   int64     `json:"periodHours"`
	PeriodStart   time.Time `json:"periodStart"`
	PeriodEnd     time.Time `json:"periodEnd"`
	Method        string    `json:"method"`
	Formula       string    `json:"formula"`
}
type BudgetAnomaly struct {
	TenantID       string    `json:"tenantId"`
	RuleID         string    `json:"ruleId"`
	Severity       string    `json:"severity"`
	DeviationCents int64     `json:"deviationCents"`
	Message        string    `json:"message"`
	ObservedAt     time.Time `json:"observedAt"`
}

type BudgetDecision struct {
	TenantID      string    `json:"tenantId"`
	Effect        string    `json:"effect"`
	Allowed       bool      `json:"allowed"`
	RuleID        string    `json:"ruleId"`
	Currency      string    `json:"currency"`
	CurrentCents  int64     `json:"currentCents"`
	ForecastCents int64     `json:"forecastCents"`
	BudgetCents   int64     `json:"budgetCents"`
	Message       string    `json:"message"`
	ObservedAt    time.Time `json:"observedAt"`
}

type BudgetNotification struct {
	TenantID   string    `json:"tenantId"`
	RuleID     string    `json:"ruleId"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	ObservedAt time.Time `json:"observedAt"`
}
