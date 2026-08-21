package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIBootstrapContextUsesOnlyAllowlistedSnapshot(t *testing.T) {
	snapshot := AIBootstrapSnapshot{TenantID: "tenant-a", ProjectID: "project-a", SessionID: "session-a", CurrentStep: 3, SessionStatus: "scanning", SCMStatus: "validated", AgentStatus: "failed", AgentError: "readiness probe failed", RunnerStatus: "waiting", ResourceScanStatus: "failed", ResourceScanError: "message password=secret", FailedCheck: "resource_scan"}
	context, err := NewAIContextBuilder(DefaultAIContextLimits()).Build(AIContextInput{TenantID: "tenant-a", Bootstrap: []AIBootstrapSnapshot{snapshot}})
	if err != nil {
		t.Fatal(err)
	}
	encodedBytes, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	encoded := strings.ToLower(string(encodedBytes))
	if strings.Contains(encoded, "registrationtoken") || strings.Contains(encoded, "secret") {
		t.Fatalf("bootstrap context contains credential-like data: %s", encoded)
	}
	if !strings.Contains(encoded, "resource_scan") || !strings.Contains(encoded, "password=[redacted]") {
		t.Fatalf("bootstrap context lost safe failure metadata: %s", encoded)
	}
}

func TestAIBootstrapContextRejectsCrossTenantSnapshot(t *testing.T) {
	_, err := NewAIContextBuilder(DefaultAIContextLimits()).Build(AIContextInput{TenantID: "tenant-a", Bootstrap: []AIBootstrapSnapshot{{TenantID: "tenant-b", SessionID: "session"}}})
	if err == nil {
		t.Fatal("cross-tenant bootstrap snapshot accepted")
	}
}
