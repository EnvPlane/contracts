package domain

import "testing"

func TestIngressCleanupIsFeatureOwnedOnly(t *testing.T) {
	endpoint := IngressEndpoint{Name: "feature-web", TLSSecretName: "feature-web-tls"}
	if !IsFeatureOwnedIngressArtifact("feature", "feature-web", endpoint) || IsFeatureOwnedIngressArtifact("feature", "shared-tls", endpoint) {
		t.Fatal("cleanup ownership guard is incorrect")
	}
}
