package domain

import "testing"

func TestCommercialSubscriptionOffers(t *testing.T) {
	c := CommercialPlanCatalog()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id                              string
		price, projects, envs, clusters int64
	}{{"community", 0, 1, 3, 1}, {"team", 9900, 5, 20, 2}, {"business", 29900, 20, 100, 5}, {"enterprise", 90000, 20, 100, 5}} {
		p, ok := CommercialPlan(tc.id)
		if !ok || p.Offer.MonthlyMinorUnits != tc.price || p.Limits[LimitProjectsMax] != tc.projects || p.Limits[LimitActiveEnvironmentsMax] != tc.envs || p.Limits[LimitManagedClustersMax] != tc.clusters {
			t.Fatalf("bad offer: %+v", p)
		}
		if p.Features[FeatureSupportSLA] || p.Features[FeatureAIDiagnosis] != (tc.price > 0) {
			t.Fatalf("incorrect feature grants: %+v", p)
		}
	}
	d := c.Deterministic()
	d.Plans[0].Offer.Name = "changed"
	if c.Plans[0].Offer.Name == "changed" {
		t.Fatal("offer copy aliases source")
	}
}
