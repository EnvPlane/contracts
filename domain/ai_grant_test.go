package domain

import (
	"testing"
	"time"
)

func TestAIAgentGrantValidation(t *testing.T) {
	now := time.Now().UTC()
	grant := AIAgentGrant{SchemaVersion: AIAgentGrantSchemaVersion, ID: "grant-1", RunID: "run-1", TenantID: "tenant-a", ProjectID: "project-a", ResourceType: "environment", ResourceID: "env-a", ToolName: "read_environment_status", Purpose: "diagnosis", RequesterID: "user-a", AgentID: "agent-1", ContextHash: "ctx", IssuedAt: now, ExpiresAt: now.Add(time.Minute)}
	if err := grant.Validate(now); err != nil {
		t.Fatal(err)
	}
	grant.ExpiresAt = now.Add(-time.Second)
	if err := grant.Validate(now); err == nil {
		t.Fatal("expired grant accepted")
	}
}
