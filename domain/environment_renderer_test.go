package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderEnvironmentReleasePlanIsDeterministicAndTransformsCMSReferences(t *testing.T) {
	revision := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev-cms", TemplateID: "cms", TenantID: "tenant-a", ProjectID: "cms", SourceScanID: "scan-cms", SourceNamespaces: []string{"dev-cms"}, Resources: []ResourceTemplate{
		{Kind: "Deployment", Namespace: "dev-cms", Name: "web", Manifest: map[string]any{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": map[string]any{"name": "web", "namespace": "dev-cms"}, "spec": map[string]any{"template": map[string]any{"spec": map[string]any{"containers": []any{map[string]any{"name": "web", "image": "ghcr.io/acme/web:old", "env": []any{map[string]any{"name": "API_URL", "value": "https://api.dev-cms.svc.cluster.local:8443/v1?q=a%2Fb"}}}}}}}}},
		{Kind: "Service", Namespace: "dev-cms", Name: "api", Manifest: map[string]any{"apiVersion": "v1", "kind": "Service", "metadata": map[string]any{"name": "api", "namespace": "dev-cms"}}},
		{Kind: "Ingress", Namespace: "dev-cms", Name: "web", Manifest: map[string]any{"apiVersion": "networking.k8s.io/v1", "kind": "Ingress", "metadata": map[string]any{"name": "web", "namespace": "dev-cms"}, "spec": map[string]any{"rules": []any{map[string]any{"host": "web.dev-cms.local"}}, "tls": []any{map[string]any{"hosts": []any{"web.dev-cms.local"}}}}}},
	}, ResourcePolicies: []ResourceDependencyPolicy{{ResourceID: "Deployment/dev-cms/web", Kind: "Deployment", Namespace: "dev-cms", Name: "web", Strategy: ResourcePolicyClone}, {ResourceID: "Service/dev-cms/api", Kind: "Service", Namespace: "dev-cms", Name: "api", Strategy: ResourcePolicyClone}, {ResourceID: "Ingress/dev-cms/web", Kind: "Ingress", Namespace: "dev-cms", Name: "web", Strategy: ResourcePolicyClone}}}
	digest, err := revision.CanonicalDigest(); if err != nil { t.Fatal(err) }; revision.Digest = digest
	graph := ServiceGraph{Nodes: []ServiceGraphNode{{ID: "Deployment/dev-cms/web", Kind: "Deployment", Namespace: "dev-cms", Name: "web"}, {ID: "Service/dev-cms/api", Kind: "Service", Namespace: "dev-cms", Name: "api"}, {ID: "Ingress/dev-cms/web", Kind: "Ingress", Namespace: "dev-cms", Name: "web"}}, Validation: &DependencyGraphValidation{Valid: true}, Policies: revision.ResourcePolicies}
	inputs := EnvironmentRenderInputs{TenantID: "tenant-a", ProjectID: "cms", EnvironmentID: "pr-42", CommitSHA: "deadbeef", NamespaceMap: map[string]string{"dev-cms": "cms-pr-42"}, ProjectDomain: "preview.example.com", ComponentImages: map[string]string{"web": "ghcr.io/acme/web:deadbeef"}, ImmutableImages: true}
	first, err := RenderEnvironmentReleasePlan(revision, graph, inputs); if err != nil { t.Fatal(err) }; second, err := RenderEnvironmentReleasePlan(revision, graph, inputs); if err != nil { t.Fatal(err) }
	if first.Digest != second.Digest || len(first.Transformations) == 0 { t.Fatalf("render is not deterministic or report is empty: %s %s", first.Digest, second.Digest) }
	encoded := ""; for _, resource := range first.RenderedResources { encoded += string(mustJSON(resource.Manifest)) }; if strings.Contains(encoded, "dev-cms") || strings.Contains(encoded, ":latest") || !strings.Contains(encoded, "cms-pr-42") || !strings.Contains(encoded, "deadbeef") || !strings.Contains(encoded, "q=a%2Fb") { t.Fatalf("rendered plan contains unresolved values: %s", encoded) }
	if first.RenderedResources[0].Name == "web" { t.Fatal("resource name was not feature-scoped") }
	for _, transformation := range first.Transformations { if transformation.FromDigest == "" || transformation.ToDigest == "" { t.Fatalf("unsafe transformation report: %#v", transformation) } }
}

func TestRenderEnvironmentReleasePlanBlocksUnresolvedRequiredLink(t *testing.T) {
	revision := EnvironmentTemplateRevision{ContractVersion: EnvironmentTemplateContractVersion, RevisionID: "rev", TemplateID: "tpl", TenantID: "tenant-a", ProjectID: "p", SourceScanID: "scan", Resources: []ResourceTemplate{{Kind: "Deployment", Namespace: "dev", Name: "web", Manifest: map[string]any{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": map[string]any{"name": "web", "namespace": "dev"}}}}, ResourcePolicies: []ResourceDependencyPolicy{{ResourceID: "Deployment/dev/web", Kind: "Deployment", Namespace: "dev", Name: "web", Strategy: ResourcePolicyClone}}}; revision.Digest, _ = revision.CanonicalDigest()
	_, err := RenderEnvironmentReleasePlan(revision, ServiceGraph{Validation: &DependencyGraphValidation{Valid: true}, Edges: []ServiceGraphEdge{{From: "Deployment/dev/web", To: "ConfigMap/dev/missing", Required: true, Path: "manifest.spec.template.spec.volumes[0]"}}, Policies: revision.ResourcePolicies}, EnvironmentRenderInputs{TenantID: "tenant-a", ProjectID: "p", EnvironmentID: "pr-1", CommitSHA: "sha", NamespaceMap: map[string]string{"dev": "pr-1"}})
	if err == nil || !strings.Contains(err.Error(), "unresolved required link") { t.Fatalf("expected unresolved link error, got %v", err) }
}

func mustJSON(value any) []byte { payload, err := json.Marshal(value); if err != nil { panic(err) }; return payload }
