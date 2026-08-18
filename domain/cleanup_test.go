package domain

import "testing"

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
