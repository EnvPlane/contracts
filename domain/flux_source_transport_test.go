package domain

import (
	"testing"
	"time"
)

func TestAgentFluxSourceTransportIsCredentialFreeAndBound(t *testing.T) {
	now := time.Unix(1, 0)
	command := AgentFluxSourceCommand{
		ContractVersion: FluxSourceCommandContractVersion,
		CommandID:       "source-1", TenantID: "tenant", ProjectID: "checkout", ClusterID: "cluster", AgentID: "agent",
		Namespace: "flux-system", GitRepositoryName: "checkout-gitops", CredentialSecretName: "checkout-gitops-auth",
		RepositoryURL: "https://github.com/envplane/gitops.git", Branch: "main", Status: FluxSourceCommandPending, CreatedAt: now,
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("validate command: %v", err)
	}
	result := AgentFluxSourceResult{ContractVersion: FluxSourceCommandContractVersion, CommandID: command.CommandID, AttemptID: "attempt", TenantID: command.TenantID, ProjectID: command.ProjectID, ClusterID: command.ClusterID, AgentID: command.AgentID, Status: FluxSourceCommandSucceeded, FinishedAt: now}
	if err := result.Validate(); err != nil {
		t.Fatalf("validate result: %v", err)
	}
}
