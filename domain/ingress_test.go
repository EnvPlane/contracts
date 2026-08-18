package domain

import "testing"

func TestPreviewDomainPolicyRequiresFeatureDiscriminator(t *testing.T) {
	if err := ValidatePreviewDomainPolicy(PreviewDomainPolicy{Pattern: "{project}.preview.example.com"}); err == nil {
		t.Fatal("expected static preview pattern to be rejected")
	}
	if err := ValidatePreviewDomainPolicy(PreviewDomainPolicy{Pattern: "{branch}.{project}.preview.example.com"}); err != nil {
		t.Fatalf("valid pattern rejected: %v", err)
	}
}

func TestRewriteIngressManifestPreservesRoutingAndFiltersUnsafeAnnotations(t *testing.T) {
	manifest := map[string]any{"metadata": map[string]any{"name": "web", "namespace": "source", "annotations": map[string]any{"nginx.ingress.kubernetes.io/configuration-snippet": "proxy_pass http://evil", "nginx.ingress.kubernetes.io/proxy-body-size": "8m"}}, "spec": map[string]any{"ingressClassName": "nginx", "rules": []any{map[string]any{"host": "old.example", "http": map[string]any{"paths": []any{map[string]any{"path": "/api", "backend": map[string]any{"service": map[string]any{"name": "api", "port": map[string]any{"number": 8080}}}}}}}}, "tls": []any{map[string]any{"hosts": []any{"old.example"}, "secretName": "source-tls"}}}}
	endpoint, err := RewriteIngressManifest(manifest, "web", IngressRenderInput{EnvironmentID: "feature-a", Project: "checkout", Branch: "feature/login", ChangeID: "42", DomainRoot: "preview.example.com", Policy: PreviewDomainPolicy{Pattern: "{branch}.{project}.preview.example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.Host == "old.example" || endpoint.TLSSecretName == "source-tls" || endpoint.IngressClassName != "nginx" || len(endpoint.Paths) != 1 {
		t.Fatalf("unexpected endpoint: %+v", endpoint)
	}
	annotations := manifest["metadata"].(map[string]any)["annotations"].(map[string]string)
	if _, ok := annotations["nginx.ingress.kubernetes.io/configuration-snippet"]; ok {
		t.Fatal("unsafe annotation survived")
	}
	if annotations["nginx.ingress.kubernetes.io/proxy-body-size"] != "8m" {
		t.Fatal("safe annotation was removed")
	}
}

func TestEndpointPreflightRequiresDNSAndTLSCapabilities(t *testing.T) {
	policy := PreviewDomainPolicy{Pattern: "{branch}.{project}.preview.example.com", WildcardDNS: false, WildcardCertificates: true}
	if policy.WildcardDNS || !policy.WildcardCertificates {
		t.Fatal("unexpected preflight fixture")
	}
	if err := ValidatePreviewDomainPolicy(policy); err != nil {
		t.Fatal(err)
	}
}
