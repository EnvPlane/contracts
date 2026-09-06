package domain

import (
	"testing"
	"time"
)

func TestEnvironmentTemplateRevisionDigestIsStableAndImmutable(t *testing.T) {
	r := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-1", TemplateID: "tpl-1", TenantID: "tenant-a", ProjectID: "project-a", SourceScanID: "scan-1", Resources: []ResourceTemplate{{Kind: "ConfigMap", Name: "app", Manifest: map[string]any{"data": map[string]any{"a": "b"}}}}}
	d1, err := r.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	d2, err := r.CanonicalDigest()
	if err != nil || d1 != d2 {
		t.Fatalf("digest is not stable: %q %q %v", d1, d2, err)
	}
	r.Digest = d1
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	r.Resources[0].Name = "changed"
	if err := r.Validate(); err == nil {
		t.Fatal("expected mutation to invalidate immutable digest")
	}
}

func TestLegacyProjectConfigRemainsReadableWithoutTemplateBinding(t *testing.T) {
	var config ProjectConfig
	if config.TemplateRevisionID != "" || config.TemplateDigest != "" {
		t.Fatal("legacy config must not acquire a synthetic template binding")
	}
}

func TestEnvironmentTemplateRejectsSecretResources(t *testing.T) {
	r := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-1", TemplateID: "tpl-1", TenantID: "tenant-a", ProjectID: "project-a", SourceScanID: "scan-1", Resources: []ResourceTemplate{{Kind: "Secret", Name: "credentials"}}}
	digest, err := r.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	r.Digest = digest
	if err := r.Validate(); err == nil {
		t.Fatal("expected Secret resource to be rejected")
	}
}

func TestEnvironmentTemplateDigestIncludesDependencyPolicies(t *testing.T) {
	r := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-cms", TemplateID: "cms", TenantID: "tenant-a", ProjectID: "cms", SourceScanID: "scan-cms", Resources: []ResourceTemplate{{Kind: "Deployment", Namespace: "dev-cms", Name: "web", Manifest: map[string]any{"apiVersion": "apps/v1"}}}, ResourcePolicies: []ResourceDependencyPolicy{{ResourceID: "ConfigMap/dev-cms/web-config", Kind: "ConfigMap", Namespace: "dev-cms", Name: "web-config", Strategy: ResourcePolicyClone, Defaulted: true, Reason: "workload-owned desired state defaults to clone"}}}
	d1, err := r.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	r.ResourcePolicies[0].Strategy = ResourcePolicyReference
	d2, err := r.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if d1 == d2 {
		t.Fatal("dependency policy changes must change revision digest")
	}
}

func TestReleasePlanExecutionInputDigestIgnoresVolatileEnvironmentFields(t *testing.T) {
	environment := Environment{ID: "env-a", Project: "project-a", Namespace: "ns-a", Charts: ChartVersions{App: "app:1"}}
	first, err := ReleasePlanExecutionInputDigest(environment, ProjectConfig{}, "chart-a", "1.0.0", 1)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(123, 0).UTC()
	environment.Status = StatusReady
	environment.LastActivityAt = &now
	environment.CostEstimate = &CostEstimate{}
	environment.Endpoints = []IngressEndpoint{{Host: "preview.example"}}
	environment.PinnedUntil = &now
	second, err := ReleasePlanExecutionInputDigest(environment, ProjectConfig{}, "chart-a", "1.0.0", 1)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("volatile Environment fields changed digest: %q != %q", first, second)
	}
}

func TestReleasePlanExecutionInputDigestIncludesDeploymentFields(t *testing.T) {
	base := Environment{ID: "env-a", Project: "project-a", Namespace: "ns-a", Charts: ChartVersions{App: "app:1"}}
	first, err := ReleasePlanExecutionInputDigest(base, ProjectConfig{}, "chart-a", "1.0.0", 1)
	if err != nil {
		t.Fatal(err)
	}
	base.Charts.App = "app:2"
	second, err := ReleasePlanExecutionInputDigest(base, ProjectConfig{}, "chart-a", "1.0.0", 1)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("deployment-relevant chart mutation did not change digest")
	}
}
