package domain

import "testing"

func TestCompareFeatureInventoryProducesSafeDrift(t *testing.T) {
	expected := []ReleasePlanInventoryItem{{Kind: "Deployment", Namespace: "feature", Name: "cms", Digest: "sha256:a"}, {Kind: "Ingress", Namespace: "feature", Name: "cms", Digest: "sha256:b"}}
	observed := FeatureInventoryReport{Complete: true, Items: []ObservedInventoryItem{{ReleasePlanInventoryItem: expected[0]}, {ReleasePlanInventoryItem: ReleasePlanInventoryItem{Kind: "ConfigMap", Namespace: "feature", Name: "extra", Digest: "sha256:x"}}}}
	diff, err := CompareFeatureInventory(expected, observed)
	if err != nil || diff.Safe || len(diff.Missing) != 1 || len(diff.Extra) != 1 {
		t.Fatalf("diff=%#v err=%v", diff, err)
	}
}

func TestFeatureReadinessIgnoresSourceDiagnosticHealth(t *testing.T) {
	observed := FeatureInventoryReport{Complete: true, Items: []ObservedInventoryItem{{ReleasePlanInventoryItem: ReleasePlanInventoryItem{Kind: "Deployment", Namespace: "feature", Name: "cms", Digest: "sha256:a"}, Health: &ResourceHealth{Status: "ready"}}}, SourceHealthDiagnostic: []ObservedInventoryItem{{Health: &ResourceHealth{Status: "unhealthy"}}}}
	diff, _ := CompareFeatureInventory([]ReleasePlanInventoryItem{{Kind: "Deployment", Namespace: "feature", Name: "cms", Digest: "sha256:a"}}, observed)
	if readiness := ValidateFeatureReadiness(true, diff, observed, true); !readiness.Ready {
		t.Fatalf("diagnostic source health blocked readiness: %#v", readiness)
	}
}

func TestIncompleteFeatureInventoryFailsClosed(t *testing.T) {
	if _, err := CompareFeatureInventory(nil, FeatureInventoryReport{}); err == nil {
		t.Fatal("incomplete inventory must fail closed")
	}
}
