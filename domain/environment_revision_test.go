package domain

import "testing"

func TestEnvironmentRevisionRejectsMutableArtifact(t *testing.T) {
	if err := (EnvironmentRevision{Commit: "abc", ArtifactDigest: "latest"}).Validate(); err == nil {
		t.Fatal("expected mutable artifact to be rejected")
	}
}

func TestEnvironmentRevisionOrdersAndDeduplicates(t *testing.T) {
	base := EnvironmentRevision{Provider: ProviderGitHub, Repository: "owner/app", ChangeID: "7", Commit: "one", Sequence: 4}
	newer := base
	newer.Commit = "two"
	newer.Sequence = 5
	if !newer.NewerThan(base) || base.NewerThan(newer) {
		t.Fatal("revision ordering is not monotonic")
	}
	if newer.NewerThan(newer) {
		t.Fatal("equal revision must be idempotent")
	}
}
