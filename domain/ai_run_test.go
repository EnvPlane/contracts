package domain

import (
	"testing"
	"time"
)

func TestAIRunTransitionsAreTerminal(t *testing.T) {
	if err := ValidateAIRunTransition(AIRunStatusSucceeded, AIRunStatusRunning); err == nil { t.Fatal("terminal run must not resume") }
	if err := ValidateAIRunTransition(AIRunStatusRunning, AIRunStatusSucceeded); err != nil { t.Fatal(err) }
	if err := ValidateAIRunTransition(AIRunStatusFailed, AIRunStatusFailed); err != nil { t.Fatal(err) }
}

func TestAIRunValidateRequiresBoundedMetadata(t *testing.T) {
	now := time.Now().UTC()
	run := AIRun{SchemaVersion: AIRunSchemaVersion, ID: "run-1", IdempotencyKey: "key-1", TenantID: "tenant-a", ProjectID: "project-a", Purpose: "diagnosis", Status: AIRunStatusQueued, Provider: "openai", Model: "gpt-5.4", PromptTemplateVersion: "diagnosis.v1", ContextHash: "sha256:x", RequestedAt: now, CreatedAt: now, UpdatedAt: now}
	if err := run.Validate(); err != nil { t.Fatal(err) }
	run.ContextHash = ""
	if err := run.Validate(); err == nil { t.Fatal("context hash is required") }
}
