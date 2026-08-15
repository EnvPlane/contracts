package domain

import "testing"

func TestParseDeploymentBackend(t *testing.T) {
	tests := []struct {
		input string
		want  DeploymentBackend
	}{
		{"helm_direct", DeploymentBackendHelmDirect},
		{"helm-direct", DeploymentBackendHelmDirect},
		{"flux", DeploymentBackendFluxCD},
		{"fluxcd", DeploymentBackendFluxCD},
		{"gitops-manifest", DeploymentBackendGitOpsManifest},
		{"argocd", DeploymentBackendArgoCD},
	}
	for _, test := range tests {
		got, err := ParseDeploymentBackend(test.input)
		if err != nil || got != test.want {
			t.Fatalf("ParseDeploymentBackend(%q) = %q, %v; want %q", test.input, got, err, test.want)
		}
	}
}

func TestParseDeploymentBackendRejectsUnknownValue(t *testing.T) {
	if _, err := ParseDeploymentBackend("custom"); err == nil {
		t.Fatal("expected unknown deployment backend to be rejected")
	}
}
