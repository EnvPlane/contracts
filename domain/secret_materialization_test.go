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
		{ID: "clone", Strategy: SecretStrategyEncryptedClone, EncryptedPayloadRef: "manualSecrets/clone", TargetName: "db-clone", TargetNamespace: "feature-a"},
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
