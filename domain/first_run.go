package domain

import (
	"fmt"
	"time"
)

// FirstRunState is the installation-scoped, restart-safe onboarding state.
// It contains no operator identity, credential material, or user-entered
// configuration so it is safe to expose through the unauthenticated progress
// endpoint.
type FirstRunState string

const (
	FirstRunUninitialized    FirstRunState = "uninitialized"
	FirstRunOperatorClaimed  FirstRunState = "operator-claimed"
	FirstRunClusterReady     FirstRunState = "cluster-ready"
	FirstRunSCMReady         FirstRunState = "scm-ready"
	FirstRunProjectReady     FirstRunState = "project-ready"
	FirstRunEnvironmentReady FirstRunState = "environment-ready"
	FirstRunComplete         FirstRunState = "complete"
)

const FirstRunProgressSchemaVersion = "v1"

// FirstRunProgress is the safe, persisted read model for the installation
// wizard. Revision provides optimistic concurrency for transitions. The claim
// is intentionally only a boolean/timestamp: the claimant is never revealed.
type FirstRunProgress struct {
	SchemaVersion    string        `json:"schemaVersion"`
	State            FirstRunState `json:"state"`
	Revision         int64         `json:"revision"`
	Claimed          bool          `json:"claimed"`
	ClaimedAt        *time.Time    `json:"claimedAt,omitempty"`
	LastTransitionAt *time.Time    `json:"lastTransitionAt,omitempty"`
	UpdatedAt        *time.Time    `json:"updatedAt,omitempty"`
}

func InitialFirstRunProgress() FirstRunProgress {
	return FirstRunProgress{SchemaVersion: FirstRunProgressSchemaVersion, State: FirstRunUninitialized}
}

func (p FirstRunProgress) Validate() error {
	if p.SchemaVersion == "" {
		return nil // Backward-compatible decode of settings written before first-run support.
	}
	if p.SchemaVersion != FirstRunProgressSchemaVersion || !p.State.Valid() || p.Revision < 0 {
		return fmt.Errorf("invalid first-run progress")
	}
	if p.Claimed != (p.State != FirstRunUninitialized) {
		return fmt.Errorf("invalid first-run claim state")
	}
	if p.Claimed && p.ClaimedAt == nil {
		return fmt.Errorf("missing first-run claim time")
	}
	return nil
}

func (s FirstRunState) Valid() bool {
	switch s {
	case FirstRunUninitialized, FirstRunOperatorClaimed, FirstRunClusterReady,
		FirstRunSCMReady, FirstRunProjectReady, FirstRunEnvironmentReady, FirstRunComplete:
		return true
	default:
		return false
	}
}

// CanAdvanceTo allows a repeated confirmed transition and exactly one next
// state. Skipping steps fails closed.
func (p FirstRunProgress) CanAdvanceTo(next FirstRunState) bool {
	if next == p.State {
		return true
	}
	switch p.State {
	case FirstRunUninitialized:
		return next == FirstRunOperatorClaimed
	case FirstRunOperatorClaimed:
		return next == FirstRunClusterReady
	case FirstRunClusterReady:
		return next == FirstRunSCMReady
	case FirstRunSCMReady:
		return next == FirstRunProjectReady
	case FirstRunProjectReady:
		return next == FirstRunEnvironmentReady
	case FirstRunEnvironmentReady:
		return next == FirstRunComplete
	default:
		return false
	}
}
