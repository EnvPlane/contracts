package domain

import (
	"errors"
	"strings"
	"time"
)

const SecretMaterializationCommandContractVersion = "v1"

type SecretMaterializationCommandStatus string

const (
	SecretCommandPending   SecretMaterializationCommandStatus = "pending"
	SecretCommandClaimed   SecretMaterializationCommandStatus = "claimed"
	SecretCommandSucceeded SecretMaterializationCommandStatus = "succeeded"
	SecretCommandFailed    SecretMaterializationCommandStatus = "failed"
)

// AgentSecretMaterializationCommand is the metadata-only wire contract between
// control-plane and the authenticated target-cluster Agent. Secret values,
// envelopes, leases, and credentials are deliberately absent.
type AgentSecretMaterializationCommand struct {
	ContractVersion  string                             `json:"contractVersion"`
	CommandID        string                             `json:"commandId"`
	TenantID         string                             `json:"tenantId"`
	ProjectID        string                             `json:"projectId"`
	EnvironmentID    string                             `json:"environmentId"`
	ClusterID        string                             `json:"clusterId"`
	AgentID          string                             `json:"agentId"`
	Operation        SecretMaterializationOperation     `json:"operation"`
	PlanID           string                             `json:"planId"`
	PlanDigest       string                             `json:"planDigest"`
	ExpectedRevision int64                              `json:"expectedRevision"`
	Plan             SecretMaterializationPlan          `json:"plan"`
	Status           SecretMaterializationCommandStatus `json:"status"`
	Attempt          int                                `json:"attempt"`
	AttemptID        string                             `json:"attemptId,omitempty"`
	CreatedAt        time.Time                          `json:"createdAt"`
	ClaimedAt        *time.Time                         `json:"claimedAt,omitempty"`
	LeaseExpiresAt   *time.Time                         `json:"leaseExpiresAt,omitempty"`
}

func (c AgentSecretMaterializationCommand) Validate() error {
	if c.ContractVersion != SecretMaterializationCommandContractVersion || strings.TrimSpace(c.CommandID) == "" || strings.TrimSpace(c.TenantID) == "" || strings.TrimSpace(c.ProjectID) == "" || strings.TrimSpace(c.EnvironmentID) == "" || strings.TrimSpace(c.ClusterID) == "" || strings.TrimSpace(c.AgentID) == "" || strings.TrimSpace(c.PlanID) == "" || strings.TrimSpace(c.PlanDigest) == "" || c.ExpectedRevision < 1 || c.CreatedAt.IsZero() {
		return errors.New("invalid secret materialization command binding")
	}
	if c.Operation != SecretOperationMaterialize && c.Operation != SecretOperationCleanup {
		return errors.New("unsupported secret materialization command operation")
	}
	if c.Status != SecretCommandPending && c.Status != SecretCommandClaimed {
		return errors.New("secret materialization command is not claimable")
	}
	if err := c.Plan.Validate(); err != nil {
		return err
	}
	if c.Plan.PlanID != c.PlanID || c.Plan.Digest != c.PlanDigest || c.Plan.TenantID != c.TenantID || c.Plan.ProjectID != c.ProjectID || c.Plan.EnvironmentID != c.EnvironmentID || c.Plan.Revision != c.ExpectedRevision {
		return errors.New("secret materialization command does not match immutable plan")
	}
	return nil
}

type AgentSecretMaterializationResult struct {
	ContractVersion  string                             `json:"contractVersion"`
	CommandID        string                             `json:"commandId"`
	AttemptID        string                             `json:"attemptId"`
	TenantID         string                             `json:"tenantId"`
	ProjectID        string                             `json:"projectId"`
	EnvironmentID    string                             `json:"environmentId"`
	ClusterID        string                             `json:"clusterId"`
	AgentID          string                             `json:"agentId"`
	PlanID           string                             `json:"planId"`
	PlanDigest       string                             `json:"planDigest"`
	ExpectedRevision int64                              `json:"expectedRevision"`
	Status           SecretMaterializationCommandStatus `json:"status"`
	ErrorCode        SecretMaterializationErrorCode     `json:"errorCode,omitempty"`
	Items            []SecretMaterializationItemResult  `json:"items,omitempty"`
	FinishedAt       time.Time                          `json:"finishedAt"`
}

func (r AgentSecretMaterializationResult) Validate() error {
	if r.ContractVersion != SecretMaterializationCommandContractVersion || strings.TrimSpace(r.CommandID) == "" || strings.TrimSpace(r.AttemptID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.EnvironmentID) == "" || strings.TrimSpace(r.ClusterID) == "" || strings.TrimSpace(r.AgentID) == "" || strings.TrimSpace(r.PlanID) == "" || strings.TrimSpace(r.PlanDigest) == "" || r.ExpectedRevision < 1 || r.FinishedAt.IsZero() {
		return errors.New("invalid secret materialization result binding")
	}
	if r.Status != SecretCommandSucceeded && r.Status != SecretCommandFailed {
		return errors.New("secret materialization result is not terminal")
	}
	if r.Status == SecretCommandSucceeded && r.ErrorCode != "" {
		return errors.New("successful secret materialization result has an error code")
	}
	return nil
}
