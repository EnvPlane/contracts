package domain

import (
	"errors"
	"strings"
	"time"
)

const AIGitOpsAgentSchemaVersion = "1"

type AIGitOpsDiagnosisCode string

const (
	AIGitOpsHealthy      AIGitOpsDiagnosisCode = "healthy"
	AIGitOpsMissingPath  AIGitOpsDiagnosisCode = "missing_path"
	AIGitOpsInvalidYAML  AIGitOpsDiagnosisCode = "invalid_yaml"
	AIGitOpsStaleSource  AIGitOpsDiagnosisCode = "stale_source"
	AIGitOpsConflict     AIGitOpsDiagnosisCode = "conflict"
	AIGitOpsUnauthorized AIGitOpsDiagnosisCode = "unauthorized"
	AIGitOpsUnknown      AIGitOpsDiagnosisCode = "unknown"
)

type AIGitOpsRepairKind string

const (
	AIGitOpsRetryReconciliation AIGitOpsRepairKind = "retry_reconciliation"
	AIGitOpsProposeManifestFix  AIGitOpsRepairKind = "propose_manifest_fix"
	AIGitOpsRollbackCommit      AIGitOpsRepairKind = "rollback_commit"
)

type AIGitOpsObservation struct {
	Repository       string `json:"repository"`
	AuthorizedPath   string `json:"authorizedPath"`
	OutputPath       string `json:"outputPath"`
	SourceRevision   string `json:"sourceRevision"`
	AppliedRevision  string `json:"appliedRevision"`
	Controller       string `json:"controller"`
	ControllerStatus string `json:"controllerStatus"`
	PathExists       bool   `json:"pathExists"`
	ManifestValid    bool   `json:"manifestValid"`
	OwnershipValid   bool   `json:"ownershipValid"`
	Conflict         bool   `json:"conflict"`
	ForcePush        bool   `json:"forcePush"`
	ProtectedBranch  bool   `json:"protectedBranch"`
}

type AIGitOpsRepairProposal struct {
	Kind             AIGitOpsRepairKind `json:"kind"`
	Repository       string             `json:"repository"`
	Branch           string             `json:"branch"`
	AuthorizedPath   string             `json:"authorizedPath"`
	Files            []string           `json:"files,omitempty"`
	ExactDiff        string             `json:"exactDiff,omitempty"`
	RollbackCommit   string             `json:"rollbackCommit,omitempty"`
	ApprovalRequired bool               `json:"approvalRequired"`
	Tool             string             `json:"tool"`
}

type AIGitOpsAgentPlan struct {
	SchemaVersion           string                   `json:"schemaVersion"`
	PlanID                  string                   `json:"planId"`
	TenantID                string                   `json:"tenantId"`
	ProjectID               string                   `json:"projectId"`
	ContextHash             string                   `json:"contextHash"`
	Observation             AIGitOpsObservation      `json:"observation"`
	Diagnosis               AIGitOpsDiagnosisCode    `json:"diagnosis"`
	Evidence                []AIEvidenceReference    `json:"evidence"`
	Proposals               []AIGitOpsRepairProposal `json:"proposals,omitempty"`
	BlockedReasons          []string                 `json:"blockedReasons,omitempty"`
	AppliedRevisionVerified bool                     `json:"appliedRevisionVerified"`
	FailClosed              bool                     `json:"failClosed"`
	GeneratedAt             time.Time                `json:"generatedAt"`
}

func (o AIGitOpsObservation) Validate() error {
	for _, value := range []string{o.Repository, o.AuthorizedPath, o.OutputPath, o.SourceRevision, o.AppliedRevision, o.Controller, o.ControllerStatus} {
		if len(value) > 1024 || strings.ContainsAny(value, "\r\n") {
			return errors.New("GitOps observation contains an invalid value")
		}
	}
	if !pathWithin(o.OutputPath, o.AuthorizedPath) {
		return errors.New("GitOps output path is outside the authorized repository path")
	}
	return nil
}

func pathWithin(path, root string) bool {
	path, root = strings.Trim(strings.TrimSpace(path), "/"), strings.Trim(strings.TrimSpace(root), "/")
	if path == "" || strings.HasPrefix(path, "../") || path == ".." || strings.Contains(path, "/../") || strings.HasPrefix(path, "/") {
		return false
	}
	if root == "" || root == "." {
		return true
	}
	return path == root || strings.HasPrefix(path, root+"/")
}

func (p AIGitOpsAgentPlan) Validate() error {
	if p.SchemaVersion != AIGitOpsAgentSchemaVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.ContextHash) == "" || p.GeneratedAt.IsZero() || len(p.Evidence) == 0 {
		return errors.New("GitOps agent plan identity is invalid")
	}
	if err := p.Observation.Validate(); err != nil {
		return err
	}
	for _, evidence := range p.Evidence {
		if evidence.TenantID != p.TenantID || strings.TrimSpace(evidence.SourceType) == "" || strings.TrimSpace(evidence.SourceID) == "" {
			return errors.New("GitOps evidence is outside tenant scope")
		}
	}
	for _, proposal := range p.Proposals {
		if strings.TrimSpace(proposal.Repository) == "" || strings.TrimSpace(proposal.Branch) == "" || !pathWithin(proposal.AuthorizedPath, p.Observation.AuthorizedPath) || strings.TrimSpace(proposal.Tool) == "" || len(proposal.ExactDiff) > 65536 || strings.ContainsAny(proposal.ExactDiff, "\x00") || (proposal.Kind == AIGitOpsProposeManifestFix && len(proposal.ExactDiff) == 0) {
			return errors.New("GitOps repair proposal is invalid")
		}
		for _, file := range proposal.Files {
			if len(file) > 1024 || strings.ContainsAny(file, "\r\n\x00") || !pathWithin(file, proposal.AuthorizedPath) {
				return errors.New("GitOps repair proposal contains an unauthorized file")
			}
		}
		if proposal.Kind == AIGitOpsProposeManifestFix && !proposal.ApprovalRequired {
			return errors.New("manifest repair requires approval")
		}
		if proposal.Kind == AIGitOpsRollbackCommit && strings.TrimSpace(proposal.RollbackCommit) == "" {
			return errors.New("rollback proposal requires a rollback commit")
		}
		if p.Observation.ProtectedBranch && !proposal.ApprovalRequired {
			return errors.New("protected branch repair requires approval")
		}
	}
	if p.FailClosed && len(p.BlockedReasons) == 0 {
		return errors.New("fail-closed GitOps plan needs a reason")
	}
	return nil
}
