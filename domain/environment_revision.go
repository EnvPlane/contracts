package domain

import (
	"fmt"
	"regexp"
	"strings"
)

type EnvironmentRevision struct {
	Provider       Provider `json:"provider,omitempty"`
	Repository     string   `json:"repository,omitempty"`
	ChangeID       string   `json:"changeId,omitempty"`
	Commit         string   `json:"commit,omitempty"`
	ArtifactDigest string   `json:"artifactDigest,omitempty"`
	Sequence       int64    `json:"sequence,omitempty"`
}

var immutableDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (r EnvironmentRevision) Validate() error {
	if strings.TrimSpace(r.Commit) == "" && strings.TrimSpace(r.ArtifactDigest) == "" {
		return fmt.Errorf("revision requires commit or artifact digest")
	}
	if r.ArtifactDigest != "" && !immutableDigestPattern.MatchString(strings.ToLower(strings.TrimSpace(r.ArtifactDigest))) {
		return fmt.Errorf("artifact digest must be an immutable sha256 digest")
	}
	return nil
}

func (r EnvironmentRevision) HasImmutableDigest() bool {
	return immutableDigestPattern.MatchString(strings.ToLower(strings.TrimSpace(r.ArtifactDigest)))
}

func (r EnvironmentRevision) NewerThan(other EnvironmentRevision) bool {
	// GitLab merge-request webhooks do not provide a monotonic source-revision
	// sequence. Treat a changed immutable revision as a reconciliation signal in
	// that case: ordering commits lexicographically drops legitimate updates
	// whenever the new SHA happens to sort before the old one.
	if r.Sequence == 0 && other.Sequence == 0 {
		return r != other
	}
	if r.Sequence != other.Sequence {
		if r.Sequence == 0 {
			return false
		}
		return r.Sequence > other.Sequence
	}
	left := strings.Join([]string{string(r.Provider), r.Repository, r.ChangeID, r.Commit, r.ArtifactDigest}, "\x00")
	right := strings.Join([]string{string(other.Provider), other.Repository, other.ChangeID, other.Commit, other.ArtifactDigest}, "\x00")
	return left > right
}
