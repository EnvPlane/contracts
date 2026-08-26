package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAgentSecretMaterializationCommandIsVersionedBoundAndRedacted(t *testing.T) {
	plan, err := CompileSecretMaterializationPlan("tenant", "project", "environment", "revision", "sha256:template", "feature", []SecretStrategyConfig{{ID: "registry", Strategy: SecretStrategyEncryptedClone, SourceNamespace: "base", SourceName: "registry", TargetNamespace: "feature", TargetName: "registry", EncryptedPayloadRef: "envelopes/registry"}}, "sha256:input", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	command := AgentSecretMaterializationCommand{ContractVersion: SecretMaterializationCommandContractVersion, CommandID: "command", TenantID: plan.TenantID, ProjectID: plan.ProjectID, EnvironmentID: plan.EnvironmentID, ClusterID: "cluster", AgentID: "agent", Operation: SecretOperationMaterialize, PlanID: plan.PlanID, PlanDigest: plan.Digest, ExpectedRevision: plan.Revision, Plan: plan, Status: SecretCommandPending, Attempt: 0, CreatedAt: time.Unix(2, 0)}
	if err := command.Validate(); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"ciphertext", "leaseToken", "secretValue", "dockerconfigjson"} {
		if strings.Contains(strings.ToLower(string(payload)), strings.ToLower(forbidden)) {
			t.Fatalf("command leaked forbidden field %q", forbidden)
		}
	}
	var roundTrip AgentSecretMaterializationCommand
	if err := json.Unmarshal(payload, &roundTrip); err != nil || roundTrip.PlanDigest != plan.Digest {
		t.Fatalf("round trip failed: %v %#v", err, roundTrip)
	}
	roundTrip.TenantID = "other"
	if err := roundTrip.Validate(); err == nil {
		t.Fatal("cross-tenant plan substitution was accepted")
	}
}

func TestAgentSecretMaterializationResultRequiresTerminalBoundMetadata(t *testing.T) {
	result := AgentSecretMaterializationResult{ContractVersion: SecretMaterializationCommandContractVersion, CommandID: "command", AttemptID: "attempt", TenantID: "tenant", ProjectID: "project", EnvironmentID: "environment", ClusterID: "cluster", AgentID: "agent", PlanID: "plan", PlanDigest: "sha256:digest", ExpectedRevision: 1, Status: SecretCommandFailed, ErrorCode: SecretErrorSourceNotFound, FinishedAt: time.Unix(3, 0)}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
	result.Status = SecretCommandSucceeded
	if err := result.Validate(); err == nil {
		t.Fatal("successful result retained an error code")
	}
}
