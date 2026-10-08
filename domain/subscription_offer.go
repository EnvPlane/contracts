package domain

import "fmt"

const SubscriptionOfferVersion = "2.0.0"

// SubscriptionOffer describes a software subscription, never infrastructure
// charges or a payment authorization. Enterprise pricing is a quote floor.
type SubscriptionOffer struct {
	Name              string `json:"name"`
	Currency          string `json:"currency"`
	MonthlyMinorUnits int64  `json:"monthlyMinorUnits"`
	ContactSales      bool   `json:"contactSales"`
	CustomerAIKey     bool   `json:"customerAIKey"`
}

func (o SubscriptionOffer) Validate() error {
	if o.Name == "" || o.Currency != "EUR" || o.MonthlyMinorUnits < 0 {
		return fmt.Errorf("invalid subscription offer")
	}
	return nil
}

// CommercialPlanCatalog is the current offer. CommunityFreePlanCatalog remains
// immutable so previously issued subscriptions can retain their exact version.
func CommercialPlanCatalog() PlanCatalog {
	community := subscriptionPlan("community", "Community", 0, 1, 3, 1, false)
	free := subscriptionPlan("free", "Community", 0, 1, 3, 1, false)
	team := subscriptionPlan("team", "Team", 9900, 5, 20, 2, true)
	business := subscriptionPlan("business", "Business", 29900, 20, 100, 5, true)
	enterprise := subscriptionPlan("enterprise", "Enterprise", 90000, 20, 100, 5, true)
	enterprise.Offer.ContactSales = true
	for _, p := range []*PlanDefinition{&business, &enterprise} {
		p.Features[FeatureAuthSAML] = true
		p.Features[FeatureRBACGranular] = true
		p.Features[FeatureAuditSIEM] = true
	}
	enterprise.Features[FeatureIdentitySCIM] = true
	enterprise.Features[FeatureFleetUpgradeWaves] = true
	// SLA is granted only by an explicit negotiated license, not by a price label.
	return PlanCatalog{SchemaVersion: PlanSchemaVersion, Plans: []PlanDefinition{community, free, team, business, enterprise}}
}

func subscriptionPlan(id, name string, price, projects, environments, clusters int64, paid bool) PlanDefinition {
	p := freePlan(SubscriptionOfferVersion)
	p.ID = id
	p.Offer = &SubscriptionOffer{Name: name, Currency: "EUR", MonthlyMinorUnits: price, CustomerAIKey: paid}
	for key, value := range map[string]int64{LimitProjectsMax: projects, "maxProjects": projects, LimitActiveEnvironmentsMax: environments, "maxActiveEnvironments": environments, LimitManagedClustersMax: clusters, "maxRemoteClusters": clusters} {
		p.Limits[key] = value
	}
	p.Features[FeatureFinOpsAllocation] = paid
	p.Features[FeaturePolicyCustom] = paid
	for _, key := range []string{FeatureAIDiagnosis, FeatureAIBootstrap, FeatureAIConfiguration, FeatureAIEnvironment, FeatureAIFinOps, FeatureAIApproved, FeatureAIGitOps, FeatureAIKubernetes, FeatureAISCM, FeatureAIRelease, FeatureAIIncident, FeatureAISecurity} {
		p.Features[key] = paid
	}
	if paid {
		p.Limits[LimitAIRunsConcurrent] = 4
		p.Limits[LimitAIRunsRequests] = 10000
		p.Limits[LimitAIContextBytes] = 1048576
		p.Limits[LimitAIOutputTokens] = 16384
		p.Limits[LimitAuditRetentionDays] = 30
	}
	return p
}

func CommercialPlan(id string) (PlanDefinition, bool) {
	for _, p := range CommercialPlanCatalog().Plans {
		if p.ID == id {
			return p, true
		}
	}
	return PlanDefinition{}, false
}
