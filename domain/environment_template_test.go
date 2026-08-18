package domain

import "testing"

func TestEnvironmentTemplateRevisionDigestIsStableAndImmutable(t *testing.T) {
	r := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-1", TemplateID: "tpl-1", TenantID: "tenant-a", ProjectID: "project-a", SourceScanID: "scan-1", Resources: []ResourceTemplate{{Kind: "ConfigMap", Name: "app", Manifest: map[string]any{"data": map[string]any{"a": "b"}}}}}
	d1, err := r.CanonicalDigest(); if err != nil { t.Fatal(err) }; d2, err := r.CanonicalDigest(); if err != nil || d1 != d2 { t.Fatalf("digest is not stable: %q %q %v", d1, d2, err) }; r.Digest = d1; if err := r.Validate(); err != nil { t.Fatal(err) }; r.Resources[0].Name = "changed"; if err := r.Validate(); err == nil { t.Fatal("expected mutation to invalidate immutable digest") }
}

func TestLegacyProjectConfigRemainsReadableWithoutTemplateBinding(t *testing.T) { var config ProjectConfig; if config.TemplateRevisionID != "" || config.TemplateDigest != "" { t.Fatal("legacy config must not acquire a synthetic template binding") } }

func TestEnvironmentTemplateRejectsSecretResources(t *testing.T) {
	r := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-1", TemplateID: "tpl-1", TenantID: "tenant-a", ProjectID: "project-a", SourceScanID: "scan-1", Resources: []ResourceTemplate{{Kind: "Secret", Name: "credentials"}}}
	digest, err := r.CanonicalDigest(); if err != nil { t.Fatal(err) }; r.Digest = digest; if err := r.Validate(); err == nil { t.Fatal("expected Secret resource to be rejected") }
}

func TestEnvironmentTemplateDigestIncludesDependencyPolicies(t *testing.T) {
	r := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-cms", TemplateID: "cms", TenantID: "tenant-a", ProjectID: "cms", SourceScanID: "scan-cms", Resources: []ResourceTemplate{{Kind: "Deployment", Namespace: "dev-cms", Name: "web", Manifest: map[string]any{"apiVersion": "apps/v1"}}}, ResourcePolicies: []ResourceDependencyPolicy{{ResourceID: "ConfigMap/dev-cms/web-config", Kind: "ConfigMap", Namespace: "dev-cms", Name: "web-config", Strategy: ResourcePolicyClone, Defaulted: true, Reason: "workload-owned desired state defaults to clone"}}}
	d1, err := r.CanonicalDigest(); if err != nil { t.Fatal(err) }; r.ResourcePolicies[0].Strategy = ResourcePolicyReference; d2, err := r.CanonicalDigest(); if err != nil { t.Fatal(err) }; if d1 == d2 { t.Fatal("dependency policy changes must change revision digest") }
}
