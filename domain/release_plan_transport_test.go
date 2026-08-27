package domain

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"
)

func TestReleasePlanTransportVerifiesSignatureAndExactInventory(t *testing.T) {
	plan := EnvironmentReleasePlan{ContractVersion: EnvironmentTemplateContractVersion, PlanID: "rev-1/env-a", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", TemplateRevisionID: "rev-1", TemplateDigest: "sha256:template", InputDigest: "sha256:input", RenderedResources: []RenderedResource{{ResourceID: "Deployment/feature-a/api", Kind: "Deployment", Namespace: "feature-a", Name: "api", Digest: "sha256:resource", Manifest: map[string]any{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": map[string]any{"name": "api", "namespace": "feature-a"}}}}, Ownership: []OwnershipRecord{{Kind: "Deployment", Namespace: "feature-a", Name: "api"}}}
	plan.Digest, _ = plan.CanonicalDigest()
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	ref, err := SignReleasePlanReference(plan, "release-key-1", private)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyReleasePlanReference(plan, ref, public, ReleasePlanRunnerIdentity{TenantID: "tenant-a", ProjectID: "project-a"}, []string{"feature-a"}, []string{"Deployment"}); err != nil {
		t.Fatal(err)
	}
	if err := VerifyReleasePlanReference(plan, ref, public, ReleasePlanRunnerIdentity{TenantID: "tenant-b", ProjectID: "project-a"}, []string{"feature-a"}, []string{"Deployment"}); err == nil {
		t.Fatal("tenant substitution must be rejected")
	}
	if err := plan.ReadyFromInventory(plan.Inventory()); err != nil {
		t.Fatal(err)
	}
}
func TestReleasePlanTransportRejectsInlineCredentialsAndIncompleteOwnership(t *testing.T) {
	plan := EnvironmentReleasePlan{ContractVersion: EnvironmentTemplateContractVersion, PlanID: "rev-1/env-a", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", TemplateRevisionID: "rev-1", TemplateDigest: "sha256:template", InputDigest: "sha256:input", RenderedResources: []RenderedResource{{Kind: "ConfigMap", Namespace: "feature-a", Name: "config", Digest: "sha256:resource", Manifest: map[string]any{"kind": "ConfigMap", "metadata": map[string]any{"name": "config", "namespace": "feature-a"}, "data": map[string]any{"password": "raw"}}}}}
	plan.Digest, _ = plan.CanonicalDigest()
	if err := plan.ValidateForExecution(ReleasePlanRunnerIdentity{TenantID: "tenant-a", ProjectID: "project-a"}, []string{"feature-a"}, []string{"ConfigMap"}); err == nil {
		t.Fatal("incomplete ownership/inline credential plan must be rejected")
	}
}

func TestEquivalentReleasePlansRequireSameCompleteInventory(t *testing.T) {
	plan := EnvironmentReleasePlan{ContractVersion: EnvironmentTemplateContractVersion, TenantID: "tenant", ProjectID: "project", EnvironmentID: "env", RenderedResources: []RenderedResource{{Kind: "Service", Namespace: "feature-a", Name: "api", Digest: "sha256:api"}}}
	if err := EquivalentReleasePlans(plan, plan); err != nil {
		t.Fatal(err)
	}
	other := plan
	other.RenderedResources = append(other.RenderedResources, RenderedResource{Kind: "Deployment", Namespace: "feature-a", Name: "api", Digest: "sha256:deployment"})
	if err := EquivalentReleasePlans(plan, other); err == nil {
		t.Fatal("partial backend inventory must not be equivalent")
	}
}

func TestRunnerReleasePlanTransportRoundTrip(t *testing.T) {
	plan := EnvironmentReleasePlan{ContractVersion: EnvironmentTemplateContractVersion, PlanID: "rev-1/env-a", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", TemplateRevisionID: "rev-1", TemplateDigest: "sha256:template", InputDigest: "sha256:input"}
	plan.Digest, _ = plan.CanonicalDigest()
	command := RunnerCommand{ID: "rendered-create", Operation: "create", ReleasePlanTransportVersion: ReleasePlanTransportVersion, ReleasePlanID: plan.PlanID, ReleasePlanDigest: plan.Digest, ReleasePlanSignature: "signature", ReleasePlanKeyID: "key-1", ReleasePlanPublicKey: "public-key", ReleasePlan: &plan}
	payload, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	var decoded RunnerCommand
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ReleasePlan == nil || decoded.ReleasePlan.Digest != plan.Digest || decoded.ReleasePlanTransportVersion != ReleasePlanTransportVersion {
		t.Fatalf("release plan transport did not round-trip: %#v", decoded)
	}
}
