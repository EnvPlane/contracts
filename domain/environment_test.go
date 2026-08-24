package domain

import "testing"

func TestEnvironmentGitOpsDirectoryUsesConfiguredOutputPath(t *testing.T) {
	environment := Environment{
		Project: "cms",
		ID:      "pr-123",
		GitOps:  GitOpsTarget{OutputPath: "environments/123/"},
	}

	if got, want := environment.GitOpsDirectory(), "environments/123"; got != want {
		t.Fatalf("GitOpsDirectory() = %q, want %q", got, want)
	}
	if got, want := environment.ManifestFilename(), "environments/123/flux-kustomization.yaml"; got != want {
		t.Fatalf("ManifestFilename() = %q, want %q", got, want)
	}
}

func TestEnvironmentGitOpsDirectoryKeepsLegacyFallback(t *testing.T) {
	environment := Environment{Project: "CMS", ID: "pr-123"}

	if got, want := environment.GitOpsDirectory(), "feature-envs/cms/pr-123"; got != want {
		t.Fatalf("GitOpsDirectory() = %q, want %q", got, want)
	}
}
