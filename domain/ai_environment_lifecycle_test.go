package domain

import (
	"testing"
	"time"
)

func TestAIEnvironmentLifecyclePlanRequiresApprovalForStateChanges(t *testing.T) {
	request := AIEnvironmentLifecycleRequest{SchemaVersion: AIEnvironmentLifecycleSchemaVersion, ID: "req-1", IdempotencyKey: "idem-1", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", Action: AIEnvironmentResize, ProposalID: "proposal-1", ContextHash: "ctx-1", RequestedBy: "user-a", CreatedAt: time.Unix(1, 0).UTC()}
	if !request.Action.RequiresApproval() {
		t.Fatal("resize must require approval")
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAIEnvironmentLifecyclePlanFailsClosedBeforeApprovalOrReadiness(t *testing.T) {
	request := AIEnvironmentLifecycleRequest{SchemaVersion: AIEnvironmentLifecycleSchemaVersion, ID: "req-1", IdempotencyKey: "idem-1", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", Action: AIEnvironmentRepair, ContextHash: "ctx-1", RequestedBy: "user-a", CreatedAt: time.Unix(1, 0).UTC()}
	plan, err := buildLifecycleTestPlan(request, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateForExecution(true); err != ErrAIEnvironmentLifecycleNotReady {
		t.Fatalf("not-ready execution error = %v", err)
	}
	plan, err = buildLifecycleTestPlan(request, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateForExecution(false); err != ErrAIEnvironmentLifecycleApproval {
		t.Fatalf("unapproved execution error = %v", err)
	}
}

func buildLifecycleTestPlan(request AIEnvironmentLifecycleRequest, ready bool) (AIEnvironmentLifecyclePlan, error) {
	checks := []AIEnvironmentReadinessCheck{{ID: "project", Passed: ready, Summary: "project check"}}
	plan := AIPlan{SchemaVersion: AIPlanSchemaVersion, ID: "plan-1", IdempotencyKey: request.IdempotencyKey, TenantID: request.TenantID, ProjectID: request.ProjectID, SubjectType: "environment", SubjectID: request.EnvironmentID, Purpose: "environment_lifecycle", Action: string(request.Action), ContextHash: request.ContextHash, Status: AIPlanApproved, Steps: []AIStep{{SchemaVersion: AIPlanSchemaVersion, ID: "step-1", Sequence: 1, Status: AIStepPending, Tool: AIToolCall{SchemaVersion: AIPlanSchemaVersion, ID: "tool-1", Name: "repair_environment", Arguments: map[string]string{"environmentId": request.EnvironmentID}, IdempotencyKey: "step-1", ContextHash: request.ContextHash}, Verification: AIVerification{SchemaVersion: AIPlanSchemaVersion, ID: "verify-1", Check: "ready", Expected: "yes", ReadOnly: true}}}, CreatedAt: request.CreatedAt, UpdatedAt: request.CreatedAt}
	return AIEnvironmentLifecyclePlan{SchemaVersion: AIEnvironmentLifecycleSchemaVersion, Request: request, Plan: plan, Readiness: checks, Ready: ready, ApprovalRequired: true}, nil
}

func TestAIEnvironmentLifecycleReportIncludesBoundedOutcome(t *testing.T) {
	report := AIEnvironmentLifecycleReport{SchemaVersion: AIEnvironmentLifecycleSchemaVersion, PlanID: "plan-1", TenantID: "tenant-a", ProjectID: "project-a", Status: AIPlanSucceeded, Summary: "Environment is ready.", Resources: []string{"Deployment/api"}, URL: "https://preview.example", Revision: "sha256:abc", GeneratedAt: time.Unix(1, 0).UTC()}
	if err := report.Validate(); err != nil {
		t.Fatal(err)
	}
	copyReport := report
	copyReport.Status = AIPlanExecuting
	if err := copyReport.Validate(); err == nil {
		t.Fatal("non-terminal report must be rejected")
	}
}
