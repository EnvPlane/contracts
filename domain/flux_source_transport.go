package domain

import (
	"errors"
	"strings"
	"time"
)

const FluxSourceCommandContractVersion = "v2"

type FluxSourceCommandStatus string

const (
	FluxSourceCommandPending   FluxSourceCommandStatus = "pending"
	FluxSourceCommandClaimed   FluxSourceCommandStatus = "claimed"
	FluxSourceCommandSucceeded FluxSourceCommandStatus = "succeeded"
	FluxSourceCommandFailed    FluxSourceCommandStatus = "failed"
)

// AgentFluxSourceCommand carries only the immutable, non-secret binding for a
// project-owned Flux GitRepository and its owning Kustomization. The
// credential is obtained separately by the authenticated Agent after it has
// claimed this command; it is never part of the browser API or this transport
// payload.
type AgentFluxSourceCommand struct {
	ContractVersion      string                  `json:"contractVersion"`
	CommandID            string                  `json:"commandId"`
	TenantID             string                  `json:"tenantId"`
	ProjectID            string                  `json:"projectId"`
	ClusterID            string                  `json:"clusterId"`
	AgentID              string                  `json:"agentId"`
	Namespace            string                  `json:"namespace"`
	GitRepositoryName    string                  `json:"gitRepositoryName"`
	CredentialSecretName string                  `json:"credentialSecretName"`
	KustomizationName    string                  `json:"kustomizationName"`
	KustomizationPath    string                  `json:"kustomizationPath"`
	RepositoryURL        string                  `json:"repositoryUrl"`
	Branch               string                  `json:"branch"`
	Status               FluxSourceCommandStatus `json:"status"`
	Attempt              int                     `json:"attempt"`
	AttemptID            string                  `json:"attemptId,omitempty"`
	CreatedAt            time.Time               `json:"createdAt"`
	ClaimedAt            *time.Time              `json:"claimedAt,omitempty"`
	LeaseExpiresAt       *time.Time              `json:"leaseExpiresAt,omitempty"`
}

func (c AgentFluxSourceCommand) Validate() error {
	if c.ContractVersion != FluxSourceCommandContractVersion || strings.TrimSpace(c.CommandID) == "" || strings.TrimSpace(c.TenantID) == "" || strings.TrimSpace(c.ProjectID) == "" || strings.TrimSpace(c.ClusterID) == "" || strings.TrimSpace(c.AgentID) == "" || strings.TrimSpace(c.Namespace) == "" || strings.TrimSpace(c.GitRepositoryName) == "" || strings.TrimSpace(c.CredentialSecretName) == "" || strings.TrimSpace(c.KustomizationName) == "" || strings.Trim(strings.TrimSpace(c.KustomizationPath), "/") == "" || strings.TrimSpace(c.RepositoryURL) == "" || strings.TrimSpace(c.Branch) == "" || c.CreatedAt.IsZero() {
		return errors.New("invalid Flux source command binding")
	}
	path := strings.Trim(strings.TrimSpace(c.KustomizationPath), "/")
	if strings.HasPrefix(path, "../") || strings.Contains(path, "/../") || path == ".." {
		return errors.New("invalid Flux kustomization path")
	}
	if c.Status != FluxSourceCommandPending && c.Status != FluxSourceCommandClaimed {
		return errors.New("flux source command is not claimable")
	}
	return nil
}

type AgentFluxSourceResult struct {
	ContractVersion string                  `json:"contractVersion"`
	CommandID       string                  `json:"commandId"`
	AttemptID       string                  `json:"attemptId"`
	TenantID        string                  `json:"tenantId"`
	ProjectID       string                  `json:"projectId"`
	ClusterID       string                  `json:"clusterId"`
	AgentID         string                  `json:"agentId"`
	Status          FluxSourceCommandStatus `json:"status"`
	ErrorCode       string                  `json:"errorCode,omitempty"`
	FinishedAt      time.Time               `json:"finishedAt"`
}

func (r AgentFluxSourceResult) Validate() error {
	if r.ContractVersion != FluxSourceCommandContractVersion || strings.TrimSpace(r.CommandID) == "" || strings.TrimSpace(r.AttemptID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.ClusterID) == "" || strings.TrimSpace(r.AgentID) == "" || r.FinishedAt.IsZero() {
		return errors.New("invalid Flux source result binding")
	}
	if r.Status != FluxSourceCommandSucceeded && r.Status != FluxSourceCommandFailed {
		return errors.New("flux source result is not terminal")
	}
	if r.Status == FluxSourceCommandSucceeded && strings.TrimSpace(r.ErrorCode) != "" {
		return errors.New("successful Flux source result has an error code")
	}
	return nil
}
