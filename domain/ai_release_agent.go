package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const AIReleaseAgentSchemaVersion = "1"

var aiReleaseDigestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

type AIReleaseDiagnosisCode string

const (
	AIReleaseHealthy       AIReleaseDiagnosisCode = "healthy"
	AIReleaseFailedBuild   AIReleaseDiagnosisCode = "failed_build"
	AIReleaseStaleChart    AIReleaseDiagnosisCode = "stale_chart"
	AIReleaseMissingImage  AIReleaseDiagnosisCode = "missing_image"
	AIReleaseContractDrift AIReleaseDiagnosisCode = "contract_drift"
	AIReleaseAtomicFailure AIReleaseDiagnosisCode = "atomic_release_failure"
	AIReleaseUnknown       AIReleaseDiagnosisCode = "unknown"
)

type AIReleaseArtifact struct {
	Name       string `json:"name"`
	Kind       string `json:"kind,omitempty"`
	Repository string `json:"repository,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Version    string `json:"version,omitempty"`
	Present    bool   `json:"present"`
	Compatible bool   `json:"compatible"`
}
type AIReleaseObservation struct {
	TenantID                 string              `json:"tenant_id"`
	UmbrellaVersion          string              `json:"umbrella_version"`
	CompatibilityFingerprint string              `json:"compatibility_fingerprint,omitempty"`
	Artifacts                []AIReleaseArtifact `json:"artifacts"`
	WorkflowState            string              `json:"workflow_state,omitempty"`
	ContractDrift            bool                `json:"contract_drift"`
	AtomicVerified           bool                `json:"atomic_verified"`
}
type AIReleaseRepairKind string

const (
	AIReleaseRerunWorkflow      AIReleaseRepairKind = "rerun_allowlisted_workflow"
	AIReleaseProposeArtifactFix AIReleaseRepairKind = "propose_artifact_fix"
	AIReleasePublishUmbrella    AIReleaseRepairKind = "publish_umbrella_release"
)

type AIReleaseRepairProposal struct {
	Kind             AIReleaseRepairKind `json:"kind"`
	Workflow         string              `json:"workflow,omitempty"`
	Repository       string              `json:"repository"`
	Branch           string              `json:"branch"`
	Files            []string            `json:"files,omitempty"`
	ExactDiff        string              `json:"exact_diff,omitempty"`
	ApprovalRequired bool                `json:"approval_required"`
	Tests            []string            `json:"tests"`
	RollbackVersion  string              `json:"rollback_version"`
	DigestRequired   bool                `json:"digest_required"`
}
type AIReleaseAgentPlan struct {
	SchemaVersion  string                    `json:"schema_version"`
	PlanID         string                    `json:"plan_id"`
	TenantID       string                    `json:"tenant_id"`
	ContextHash    string                    `json:"context_hash"`
	Observation    AIReleaseObservation      `json:"observation"`
	Diagnosis      AIReleaseDiagnosisCode    `json:"diagnosis"`
	Evidence       []AIEvidenceReference     `json:"evidence"`
	Proposals      []AIReleaseRepairProposal `json:"proposals"`
	BlockedReasons []string                  `json:"blocked_reasons,omitempty"`
	FailClosed     bool                      `json:"fail_closed"`
	GeneratedAt    time.Time                 `json:"generated_at"`
}
type AIReleaseAgentRequest struct {
	UmbrellaVersion          string              `json:"umbrella_version"`
	CompatibilityFingerprint string              `json:"compatibility_fingerprint,omitempty"`
	Artifacts                []AIReleaseArtifact `json:"artifacts"`
	WorkflowState            string              `json:"workflow_state,omitempty"`
	ContractDrift            bool                `json:"contract_drift"`
	AtomicVerified           bool                `json:"atomic_verified"`
	Repository               string              `json:"repository"`
	Branch                   string              `json:"branch"`
}

func (o AIReleaseObservation) Validate() error {
	if strings.TrimSpace(o.TenantID) == "" || strings.TrimSpace(o.UmbrellaVersion) == "" {
		return errors.New("release observation identity is required")
	}
	for _, artifact := range o.Artifacts {
		if strings.TrimSpace(artifact.Name) == "" || strings.ContainsAny(artifact.Name+artifact.Repository+artifact.Version, "\r\n\x00") {
			return errors.New("release artifact identity is invalid")
		}
		if !artifact.Present {
			continue
		}
		repository := strings.ToLower(strings.TrimSpace(artifact.Repository))
		lastSlash := strings.LastIndex(repository, "/")
		if strings.Contains(repository, "latest") || strings.Contains(repository, "@") || strings.Contains(repository[lastSlash+1:], ":") || !aiReleaseDigestPattern.MatchString(strings.ToLower(artifact.Digest)) {
			return errors.New("release artifact must use an immutable digest")
		}
	}
	return nil
}
func (p AIReleaseAgentPlan) Validate() error {
	if p.SchemaVersion != AIReleaseAgentSchemaVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ContextHash) == "" || p.GeneratedAt.IsZero() || len(p.Evidence) == 0 {
		return errors.New("release agent plan identity is invalid")
	}
	if err := p.Observation.Validate(); err != nil {
		return err
	}
	for _, e := range p.Evidence {
		if e.TenantID != p.TenantID || e.Validate() != nil {
			return errors.New("release evidence is outside tenant scope")
		}
	}
	for _, proposal := range p.Proposals {
		if strings.TrimSpace(proposal.Repository) == "" || strings.TrimSpace(proposal.Branch) == "" || len(proposal.Tests) == 0 || strings.TrimSpace(proposal.RollbackVersion) == "" || !proposal.DigestRequired {
			return errors.New("release proposal lacks bounded verification")
		}
		if proposal.Kind == AIReleasePublishUmbrella && !proposal.ApprovalRequired {
			return errors.New("umbrella publication requires approval")
		}
	}
	if p.FailClosed && len(p.BlockedReasons) == 0 {
		return errors.New("fail-closed release plan needs a reason")
	}
	return nil
}

func (r AIReleaseAgentRequest) Validate() error {
	if strings.TrimSpace(r.UmbrellaVersion) == "" || strings.TrimSpace(r.Repository) == "" || strings.TrimSpace(r.Branch) == "" {
		return errors.New("release agent request scope is required")
	}
	if len(r.Artifacts) > 128 {
		return errors.New("release agent artifact limit exceeded")
	}
	return AIReleaseObservation{TenantID: "request", UmbrellaVersion: r.UmbrellaVersion, CompatibilityFingerprint: r.CompatibilityFingerprint, Artifacts: r.Artifacts, WorkflowState: r.WorkflowState, ContractDrift: r.ContractDrift, AtomicVerified: r.AtomicVerified}.Validate()
}
