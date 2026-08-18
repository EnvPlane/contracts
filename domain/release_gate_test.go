package domain

import "testing"

func TestReleaseGateRejectsPlaceholderAndStaleArtifacts(t *testing.T) {
	gate := validReleaseGate()
	gate.ExpectedInventory = []ReleasePlanInventoryItem{{Kind: "Deployment", Namespace: "feature", Name: "{{ .Service }}"}}
	if err := gate.Validate(); err == nil {
		t.Fatal("placeholder inventory must fail closed")
	}
	gate = validReleaseGate()
	gate.ObservedInventory.Items = []ObservedInventoryItem{{ReleasePlanInventoryItem: ReleasePlanInventoryItem{Kind: "Deployment", Namespace: "feature", Name: "cms", Digest: "stale artifact"}}}
	if err := gate.Validate(); err == nil {
		t.Fatal("stale artifact must fail closed")
	}
}

func TestReleaseGateComparesInventoryAndGraphForBothBackends(t *testing.T) {
	for _, backend := range []DeploymentBackend{DeploymentBackendHelmDirect, DeploymentBackendFluxCD} {
		gate := validReleaseGate()
		gate.Backend = backend
		if err := gate.Validate(); err != nil {
			t.Fatalf("backend=%s: %v", backend, err)
		}
		gate.ObservedGraph = []ObservedDependencyEdge{{From: "a", To: "b", Type: "service", Required: true}}
		gate.ExpectedGraph = []ObservedDependencyEdge{{From: "a", To: "c", Type: "service", Required: true}}
		if err := gate.Validate(); err == nil {
			t.Fatalf("backend=%s graph drift was accepted", backend)
		}
	}
}

func validReleaseGate() ReleaseGate {
	item := ReleasePlanInventoryItem{Kind: "Deployment", Namespace: "feature", Name: "cms", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Owned: true}
	return ReleaseGate{Stage: ReleaseGateParity, Provider: ProviderGitLab, Backend: DeploymentBackendFluxCD, TenantID: "tenant", ProjectID: "cms", EnvironmentID: "feature", TemplateDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PlanDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", ExpectedInventory: []ReleasePlanInventoryItem{item}, ObservedInventory: FeatureInventoryReport{Complete: true, Items: []ObservedInventoryItem{{ReleasePlanInventoryItem: item}}}, SCMOpen: true, BackendApplied: true}
}
