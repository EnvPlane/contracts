package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSecretMaterializationPlanAllStrategiesAreTypedAndRedacted(t *testing.T) {
	configs := []SecretStrategyConfig{
		{ID: "ref", Strategy: SecretStrategyReference, SourceNamespace: "shared", SourceName: "db", TargetName: "db", TargetNamespace: "feature-a"},
		{ID: "external", Strategy: SecretStrategyExternal, ExternalSecretStore: "vault-store", ExternalKey: "projects/p/db", TargetName: "db-ext", TargetNamespace: "feature-a"},
		{ID: "clone", Strategy: SecretStrategyEncryptedClone, SourceNamespace: "shared", SourceName: "db-clone-source", EncryptedPayloadRef: "manualSecrets/clone", TargetName: "db-clone", TargetNamespace: "feature-a"},
		{ID: "manual", Strategy: SecretStrategyManual, EncryptedPayloadRef: "manualSecrets/manual", TargetName: "db-manual", TargetNamespace: "feature-a"},
		{ID: "generated", Strategy: SecretStrategyGenerated, Generator: "random-password-v1", TargetName: "db-generated", TargetNamespace: "feature-a"},
	}
	plan, err := CompileSecretMaterializationPlan("tenant-a", "project-a", "env-a", "rev-1", "sha256:revision", "feature-a", configs, "sha256:inputs", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Validate(); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(plan)
	text := string(payload)
	for _, forbidden := range []string{"manualValue", "plaintext", "ciphertext", "do-not-store"} {
		if strings.Contains(strings.ToLower(text), forbidden) {
			t.Fatalf("secret material leaked in plan: %s", forbidden)
		}
	}
	if len(plan.Ownership) != 3 {
		t.Fatalf("owned strategies = %d, want 3", len(plan.Ownership))
	}
	if plan.CanDeleteOwnedSecret("feature-a", "db") || !plan.CanDeleteOwnedSecret("feature-a", "db-generated") {
		t.Fatal("ownership cleanup boundary is incorrect")
	}
}

func TestSecretMaterializationPlanRejectsNamespaceEscapeAndPlaintext(t *testing.T) {
	_, err := CompileSecretMaterializationPlan("tenant-a", "project-a", "env-a", "rev-1", "sha256:revision", "feature-a", []SecretStrategyConfig{{ID: "manual", Strategy: SecretStrategyManual, TargetNamespace: "other", TargetName: "db", ManualValue: "do-not-store"}}, "sha256:inputs", time.Now())
	if err == nil {
		t.Fatal("expected namespace/plaintext rejection")
	}
}

func TestEncryptedCloneSourceNamespaceIsSignedAndCannotEscape(t *testing.T) {
	plan, err := CompileSecretMaterializationPlan("tenant", "project", "environment", "revision", "sha256:template", "target", []SecretStrategyConfig{{ID: "clone", Strategy: SecretStrategyEncryptedClone, SourceNamespace: "base", SourceName: "source", TargetNamespace: "target", TargetName: "clone", EncryptedPayloadRef: "envelopes/clone"}}, "sha256:input", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(plan.AllowedSourceNamespaces, "base") {
		t.Fatal("clone source namespace is not signed")
	}
	plan.Items[0].SourceNamespace = "kube-system"
	plan.Digest = ""
	plan.Digest, err = plan.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Validate(); err == nil {
		t.Fatal("namespace escape was accepted")
	}
}

func TestSecretMaterializationLifecycleContract(t *testing.T) {
	for _, state := range []SecretMaterializationPlanState{SecretPlanPending, SecretPlanApproved, SecretPlanMaterializing, SecretPlanReady, SecretPlanFailed, SecretPlanCleaning, SecretPlanDeleted} {
		if !validSecretPlanState(state) {
			t.Fatalf("state %q is not registered", state)
		}
	}
	transitions := [][2]SecretMaterializationPlanState{{SecretPlanPending, SecretPlanApproved}, {SecretPlanApproved, SecretPlanMaterializing}, {SecretPlanMaterializing, SecretPlanReady}, {SecretPlanReady, SecretPlanCleaning}, {SecretPlanCleaning, SecretPlanDeleted}, {SecretPlanFailed, SecretPlanMaterializing}}
	for _, transition := range transitions {
		if err := ValidateSecretMaterializationTransition(transition[0], transition[1]); err != nil {
			t.Fatalf("transition %q -> %q: %v", transition[0], transition[1], err)
		}
	}
	if err := ValidateSecretMaterializationTransition(SecretPlanDeleted, SecretPlanReady); err == nil {
		t.Fatal("deleted plan must not become ready")
	}
	for _, code := range []SecretMaterializationErrorCode{SecretErrorConflict, SecretErrorBackendUnavailable, SecretErrorTimeout} {
		if !code.Retryable() {
			t.Fatalf("%q must be retryable", code)
		}
	}
	if SecretErrorNamespaceEscape.Retryable() || SecretErrorPlaintextForbidden.Retryable() {
		t.Fatal("validation errors must be terminal")
	}
	if err := ValidateSecretMaterializationRevision(4, 3); err == nil {
		t.Fatal("stale revision was accepted")
	}
	if err := ValidateSecretMaterializationRevision(4, 4); err != nil {
		t.Fatal(err)
	}
}

func TestSecretMaterializationDigestAndIdempotencyAreDeterministic(t *testing.T) {
	first, err := SecretMaterializationIdempotencyKey("tenant", "project", "environment", "sha256:template", "namespace", "item", SecretOperationMaterialize)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SecretMaterializationIdempotencyKey("tenant", "project", "environment", "sha256:template", "namespace", "item", SecretOperationMaterialize)
	if err != nil || first != second {
		t.Fatalf("idempotency key is not deterministic: %q %q %v", first, second, err)
	}
	mapA := struct {
		Values map[string]string `json:"values"`
	}{Values: map[string]string{"b": "2", "a": "1"}}
	mapB := struct {
		Values map[string]string `json:"values"`
	}{Values: map[string]string{"a": "1", "b": "2"}}
	digestA, err := canonicalDigest(&mapA, func(any) {})
	if err != nil {
		t.Fatal(err)
	}
	digestB, err := canonicalDigest(&mapB, func(any) {})
	if err != nil || digestA != digestB {
		t.Fatalf("map ordering changed canonical digest: %q %q", digestA, digestB)
	}
}

func TestSecretMaterializationJSONRoundTripNeverCarriesValues(t *testing.T) {
	config := SecretStrategyConfig{ID: "manual", Strategy: SecretStrategyManual, TargetNamespace: "feature-a", TargetName: "db", EncryptedPayloadRef: "payload/manual", ManualValue: "secret-value"}
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret-value") || strings.Contains(string(encoded), "manualValue") {
		t.Fatalf("plaintext was serialized: %s", encoded)
	}
	var decoded SecretStrategyConfig
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ManualValue != "" {
		t.Fatal("plaintext survived JSON round-trip")
	}
}

func TestSecretMaterializationValidationRejectsAmbiguousAndMissingBindings(t *testing.T) {
	base := func(strategy SecretMaterializationStrategy) SecretMaterializationPlan {
		return SecretMaterializationPlan{ContractVersion: SecretMaterializationContractVersion, PlanID: "plan", TenantID: "tenant", ProjectID: "project", EnvironmentID: "environment", TemplateRevisionID: "revision", TemplateDigest: "sha256:template", TargetNamespace: "feature-a", InputDigest: "sha256:input", Digest: "sha256:digest", CreatedAt: time.Unix(1, 0).UTC(), Items: []SecretMaterializationItem{{ID: "item", Strategy: strategy, TargetNamespace: "feature-a", TargetName: "db", RetentionHours: 1}}}
	}
	for _, strategy := range []SecretMaterializationStrategy{SecretStrategyReference, SecretStrategyExternal, SecretStrategyEncryptedClone, SecretStrategyManual, SecretStrategyGenerated} {
		if err := base(strategy).Validate(); err == nil {
			t.Fatalf("strategy %q accepted missing bindings", strategy)
		}
	}
	plan := base(SecretStrategyGenerated)
	plan.Items[0].Generator = "generator-v1"
	plan.Items[0].CredentialRotation = "on_create_and_cleanup"
	plan.Items[0].Owned = true
	plan.Ownership = []OwnershipRecord{{Kind: "Secret", Namespace: "feature-a", Name: "db"}, {Kind: "Secret", Namespace: "feature-a", Name: "db"}}
	if err := plan.Validate(); err == nil {
		t.Fatal("ambiguous ownership was accepted")
	}
}

func TestSecretMaterializationStrategyActions(t *testing.T) {
	// #nosec G101 -- expected protocol action labels, not credentials.
	want := map[SecretMaterializationStrategy]string{SecretStrategyReference: "bind_existing_secret", SecretStrategyExternal: "resolve_external_secret", SecretStrategyEncryptedClone: "decrypt_and_clone_secret", SecretStrategyManual: "await_manual_secret_reference", SecretStrategyGenerated: "generate_secret"}
	for strategy, action := range want {
		if got := strategy.Action(SecretOperationMaterialize); got != action {
			t.Fatalf("strategy %q action = %q, want %q", strategy, got, action)
		}
		if got := strategy.Action(SecretOperationCleanup); got != "delete_owned_secret" {
			t.Fatalf("strategy %q cleanup action = %q", strategy, got)
		}
	}
}
