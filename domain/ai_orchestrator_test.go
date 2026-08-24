package domain

import (
	"testing"
	"time"
)

func TestAIOrchestratorRunRoundTripAndTerminalSemantics(t *testing.T) {
	now := time.Now().UTC()
	run := AIOrchestratorRun{
		SchemaVersion: AIOrchestratorSchemaVersion, ID: "run-1", IdempotencyKey: "idem-1", TenantID: "tenant-a", ProjectID: "project-a", Purpose: "approved_actions", Status: AIOrchestratorPaused,
		Plan:       AIPlan{SchemaVersion: AIPlanSchemaVersion, ID: "plan-1", IdempotencyKey: "plan-idem", TenantID: "tenant-a", ProjectID: "project-a", SubjectType: "environment", SubjectID: "env-a", Purpose: "approved_actions", Action: "environment.retry_or_refresh", ContextHash: "ctx", Status: AIPlanApproved, Steps: []AIStep{{SchemaVersion: AIPlanSchemaVersion, ID: "step-1", Sequence: 1, Status: AIStepPending, Tool: AIToolCall{SchemaVersion: AIPlanSchemaVersion, ID: "tool-1", Name: "refresh_environment_status", Arguments: map[string]string{"environmentId": "env-a"}, IdempotencyKey: "step-idem", ContextHash: "ctx"}, Verification: AIVerification{SchemaVersion: AIPlanSchemaVersion, ID: "verify-1", Check: "status", Expected: "current", ReadOnly: true}}}},
		Checkpoint: AIExecutionCheckpoint{Phase: AIOrchestratorPaused, PlanContextHash: "ctx", UpdatedAt: now}, Limits: AIOrchestratorLimits{MaxIterations: 3, MaxToolCalls: 2, MaxTokens: 10, MaxCostMicros: 10, MaxWallTimeSeconds: 30}, Deadline: now.Add(time.Minute), Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := run.Validate(); err != nil {
		t.Fatal(err)
	}
	if run.Terminal() {
		t.Fatal("paused run must remain resumable")
	}
	run.Status = AIOrchestratorSucceeded
	if !run.Terminal() {
		t.Fatal("succeeded run must be terminal")
	}
}
