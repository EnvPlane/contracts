package domain

import (
	"strings"
	"testing"
)

func TestIngressCleanupIsFeatureOwnedOnly(t *testing.T) {
	endpoint := IngressEndpoint{Name: "feature-web", TLSSecretName: strings.Join([]string{"feature-web", "tls"}, "-")}
	if !IsFeatureOwnedIngressArtifact("feature", "feature-web", endpoint) || IsFeatureOwnedIngressArtifact("feature", "shared-tls", endpoint) {
		t.Fatal("cleanup ownership guard is incorrect")
	}
}
