package domain

import (
	"reflect"
	"testing"
)

func TestCommunityFreePlanCatalogIsValidAndVersioned(t *testing.T) {
	catalog := CommunityFreePlanCatalog()
	if err := catalog.Validate(); err != nil {
		t.Fatalf("catalog validation: %v", err)
	}
	if len(catalog.Plans) != 2 || catalog.Plans[1].EffectiveVersion != "1.0.1" || catalog.Plans[1].Limits[LimitProjectsMax] != 1 || catalog.Plans[1].Limits["maxProjects"] != 1 || catalog.Plans[1].Limits[LimitActiveEnvironmentsMax] != 2 || catalog.Plans[1].Limits["maxActiveEnvironments"] != 2 {
		t.Fatalf("unexpected free catalog: %#v", catalog)
	}
	for _, key := range []string{FeatureGitOpsFlux, FeatureAuthOIDC, FeatureSupportSLA} {
		if _, ok := catalog.Plans[1].Features[key]; !ok {
			t.Fatalf("free plan missing canonical feature %q", key)
		}
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

func TestPlanCatalogRejectsUnknownKeysAndSchemaVersions(t *testing.T) {
	base := PlanDefinition{ID: "custom", SchemaVersion: PlanSchemaVersion, EffectiveVersion: "1.0.0"}
	for name, catalog := range map[string]PlanCatalog{
		"unknown feature": {SchemaVersion: PlanSchemaVersion, Plans: []PlanDefinition{{ID: base.ID, SchemaVersion: base.SchemaVersion, EffectiveVersion: base.EffectiveVersion, Features: map[string]bool{"feature.unknown": true}}}},
		"unknown limit":   {SchemaVersion: PlanSchemaVersion, Plans: []PlanDefinition{{ID: base.ID, SchemaVersion: base.SchemaVersion, EffectiveVersion: base.EffectiveVersion, Limits: map[string]int64{"limit.unknown": 1}}}},
		"unknown schema":  {SchemaVersion: "2", Plans: []PlanDefinition{base}},
	} {
		if err := catalog.Validate(); err == nil {
			t.Fatalf("%s: expected validation error", name)
		}
	}
}

func TestPlanCatalogDeterministicCopyPreservesHistory(t *testing.T) {
	catalog := CommunityFreePlanCatalog()
	original := catalog.Deterministic()
	deterministic := catalog.Deterministic()
	if !reflect.DeepEqual(original, deterministic) {
		t.Fatalf("deterministic catalog changed between calls")
	}
	deterministic.Plans[0].EffectiveVersion = "9.0.0"
	deterministic.Plans[0].Limits[LimitProjectsMax] = 999
	if original.Plans[0].EffectiveVersion == "9.0.0" || original.Plans[0].Limits[LimitProjectsMax] == 999 {
		t.Fatal("deterministic catalog must not rewrite the source plan history")
	}
}
