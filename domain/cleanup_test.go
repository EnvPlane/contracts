package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCleanupObservationSurvivesJSONRoundTrip(t *testing.T) {
	now := time.Unix(123, 0).UTC()
	state := CleanupState{Phase: CleanupWaitingFinalizer, ObservedAt: &now}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var restored CleanupState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.ObservedAt == nil || !restored.ObservedAt.Equal(now) {
		t.Fatal("cleanup observation lost")
	}
}

func TestCleanupInventoryExcludesUnownedResources(t *testing.T) {
	plan := EnvironmentReleasePlan{RenderedResources: []RenderedResource{{Kind: "Service", Namespace: "feature", Name: "api", Digest: "d"}, {Kind: "ConfigMap", Namespace: "base", Name: "shared", Digest: "b"}}, Ownership: []OwnershipRecord{{Kind: "Service", Namespace: "feature", Name: "api"}}}
	items := CleanupInventory(plan)
	if len(items) != 1 || items[0].Name != "api" {
		t.Fatalf("unexpected cleanup inventory: %+v", items)
	}
	if err := ValidateCleanupInventory(items, "tenant", "project", "feature"); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupInventoryRejectsForeignNamespace(t *testing.T) {
	if err := ValidateCleanupInventory([]ReleasePlanInventoryItem{{Kind: "Namespace", Namespace: "base", Name: "base", Owned: true}}, "tenant", "project", "feature"); err == nil {
		t.Fatal("expected foreign namespace rejection")
	}
}

func TestCleanupStateSupportsPartialFailureRecoveryAndReplay(t *testing.T) {
	state := CleanupState{}
	for _, phase := range []CleanupPhase{CleanupRequested, CleanupBackendDeleting, CleanupFailed, CleanupBackendDeleting, CleanupWaitingFinalizer, CleanupVerifiedEmpty, CleanupTerminated} {
		if err := state.Advance(phase); err != nil {
			t.Fatalf("advance %s: %v", phase, err)
		}
	}
	if err := state.Advance(CleanupTerminated); err != nil {
		t.Fatal(err)
	}
}
