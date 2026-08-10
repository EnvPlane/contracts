package domain

import "testing"

func TestCommunityFreePlanCatalogIsValidAndVersioned(t *testing.T) {
	catalog := CommunityFreePlanCatalog()
	if err := catalog.Validate(); err != nil {
		t.Fatalf("catalog validation: %v", err)
	}
	if len(catalog.Plans) != 2 || catalog.Plans[1].Limits["maxActiveEnvironments"] != 2 {
		t.Fatalf("unexpected free catalog: %#v", catalog)
	}
}

func TestPlanCatalogRejectsDuplicatePlansAndNegativeLimits(t *testing.T) {
	catalog := PlanCatalog{SchemaVersion: PlanSchemaVersion, Plans: []PlanDefinition{
		{ID: "free", SchemaVersion: PlanSchemaVersion, EffectiveVersion: "1.0.0", Limits: map[string]int64{"maxProjects": -1}},
		{ID: "free", SchemaVersion: PlanSchemaVersion, EffectiveVersion: "1.1.0"},
	}}
	if err := catalog.Validate(); err == nil {
		t.Fatal("expected invalid duplicate/negative plan catalog")
	}
}
