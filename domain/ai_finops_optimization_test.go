package domain

import (
	"testing"
	"time"
)

func TestAIFinOpsOptimizationRequiresEvidenceForKnownSavings(t *testing.T) {
	now := time.Now().UTC()
	ttl := 24
	p := AIFinOpsOptimizationPlan{SchemaVersion: AIFinOpsOptimizationSchemaVersion, TenantID: "tenant-a", ProjectID: "project-a", Currency: "EUR", PeriodStart: now.Add(-time.Hour), PeriodEnd: now, GeneratedAt: now, Forecast: AIFinOpsMoney{Currency: "EUR", Known: false}, Proposals: []AIFinOpsOptimizationProposal{{ID: "p", Kind: AIFinOpsOptimizeTTL, ProposalFields: ConfigurationProposalFields{TTLHours: &ttl}, Savings: AIFinOpsMoney{Currency: "EUR", MinorUnits: 10, Known: true}, PostVerification: []string{"measure next period"}, BillingAuthoritative: true}}}
	if err := p.Validate(); err == nil {
		t.Fatal("known savings without evidence must fail")
	}
}
