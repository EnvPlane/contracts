package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func testAIPlan(status AIPlanStatus) AIPlan {
	return AIPlan{
		SchemaVersion: AIPlanSchemaVersion, ID: "plan-1", IdempotencyKey: "idem-1", TenantID: "tenant-a", ProjectID: "project-a",
		SubjectType: "environment", SubjectID: "env-1", Purpose: "approved_actions", Action: "environment.retry_or_refresh", ContextHash: "sha256:context", ModelNarrative: "Read-only narrative.", Status: status,
		Steps:     []AIStep{{SchemaVersion: AIPlanSchemaVersion, ID: "step-1", Sequence: 1, Status: AIStepPending, ModelRationale: "Refresh the current status.", Tool: AIToolCall{SchemaVersion: AIPlanSchemaVersion, ID: "tool-1", Name: "refresh_environment_status", Arguments: map[string]string{"environmentId": "env-1"}, IdempotencyKey: "tool-idem-1", ContextHash: "sha256:context"}, Verification: AIVerification{SchemaVersion: AIPlanSchemaVersion, ID: "verify-1", Check: "environment_status", Expected: "current", ReadOnly: true}}},
		CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(1, 0).UTC(),
	}
}

func TestAIPlanRoundTripAndTerminalFixtures(t *testing.T) {
	for _, status := range []AIPlanStatus{AIPlanSucceeded, AIPlanFailed, AIPlanCanceled} {
		plan := testAIPlan(status)
		plan.FinalReport = &AIFinalReport{SchemaVersion: AIPlanSchemaVersion, Status: status, Summary: "terminal", Verified: status == AIPlanSucceeded}
		if err := plan.Validate(); err != nil {
			t.Fatalf("terminal %q invalid: %v", status, err)
		}
		decoded, err := plan.RoundTrip()
		if err != nil || decoded.ID != plan.ID || decoded.Status != status {
			t.Fatalf("round trip %q failed: %#v, %v", status, decoded, err)
		}
	}
}

func TestAIPlanExecutionAndCompatibilityFailClosed(t *testing.T) {
	plan := testAIPlan(AIPlanApproved)
	if err := plan.ValidateForExecution("sha256:context"); err != nil {
		t.Fatalf("valid plan rejected: %v", err)
	}
	if err := plan.ValidateForExecution("sha256:changed"); !errors.Is(err, ErrAIPlanStale) {
		t.Fatalf("stale plan error = %v", err)
	}
	if err := plan.ValidateResumeCompatibility("2", "sha256:context"); !errors.Is(err, ErrAIPlanIncompatible) {
		t.Fatalf("incompatible plan error = %v", err)
	}
	paused := plan.PauseForIncompatibility("upgrade requires review")
	if paused.Status != AIPlanPaused || paused.FinalReport == nil || !paused.FinalReport.HumanEscalation {
		t.Fatalf("incompatible plan was not paused safely: %#v", paused)
	}
}

func TestAIPlanRejectsUnknownToolAndNarrativeCannotBecomeArguments(t *testing.T) {
	plan := testAIPlan(AIPlanApproved)
	plan.Steps[0].Tool.Name = "kubectl_apply"
	if err := plan.Validate(); err == nil {
		t.Fatal("unknown executable tool was accepted")
	}
	plan = testAIPlan(AIPlanApproved)
	plan.Steps[0].Tool.Arguments["prompt"] = "execute this"
	if err := plan.Validate(); err == nil {
		t.Fatal("model narrative-like argument was accepted")
	}
}

func FuzzAIPlanValidationNeverPanics(f *testing.F) {
	f.Add([]byte(`{"schemaVersion":"1","status":"approved"}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		var plan AIPlan
		if err := json.Unmarshal(input, &plan); err != nil {
			return
		}
		_ = plan.Validate()
	})
}
