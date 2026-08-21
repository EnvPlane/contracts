package domain

import (
	"testing"
	"time"
)

func TestAIContractsValidateVersionedStatusesAndEvidence(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	request := AIRequest{
		SchemaVersion: AIContractVersion,
		RequestID:     "req-1",
		TenantID:      "tenant-a",
		Kind:          AIRequestKindDiagnosis,
		SubjectType:   "environment",
		SubjectID:     "env-1",
		RequestedAt:   now,
		Evidence:      []AIEvidenceReference{{SourceType: "environment_event", SourceID: "event-1", TenantID: "tenant-a", ObservedAt: now}},
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("request validation: %v", err)
	}
	for _, status := range []AIRunStatus{AIRunStatusQueued, AIRunStatusRunning, AIRunStatusSucceeded, AIRunStatusFailed, AIRunStatusCanceled} {
		if err := (AIRunStatusRecord{SchemaVersion: AIContractVersion, RunID: "run-1", RequestID: request.RequestID, TenantID: request.TenantID, Status: status, UpdatedAt: now}).Validate(); err != nil {
			t.Fatalf("status %q validation: %v", status, err)
		}
	}
}

func TestAIResultRequiresEvidenceForConfidence(t *testing.T) {
	valid := AIDiagnosisResult{
		SchemaVersion: AIContractVersion,
		RequestID:     "req-1",
		TenantID:      "tenant-a",
		Outcome:       AIDiagnosisOutcomeDiagnosis,
		Confidence:    AIConfidence{Score: 0.8, Level: AIConfidenceHigh, EvidenceCount: 1},
		Evidence:      []AIEvidenceReference{{SourceType: "job", SourceID: "job-1", TenantID: "tenant-a"}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid result rejected: %v", err)
	}
	valid.Evidence = nil
	valid.Confidence.EvidenceCount = 0
	if err := valid.Validate(); err == nil {
		t.Fatal("diagnosis without evidence must be rejected")
	}
	insufficient := AIDiagnosisResult{SchemaVersion: AIContractVersion, RequestID: "req-1", TenantID: "tenant-a", Outcome: AIDiagnosisOutcomeInsufficientEvidence, Confidence: AIConfidence{Level: AIConfidenceUnknown}}
	if err := insufficient.Validate(); err != nil {
		t.Fatalf("insufficient evidence result rejected: %v", err)
	}
}

func TestAIContractsRejectCrossTenantEvidenceAndUnknownErrors(t *testing.T) {
	request := AIRequest{SchemaVersion: AIContractVersion, RequestID: "req-1", TenantID: "tenant-a", Kind: AIRequestKindDiagnosis, SubjectType: "environment", SubjectID: "env-1", Evidence: []AIEvidenceReference{{SourceType: "job", SourceID: "job-1", TenantID: "tenant-b"}}}
	if err := request.Validate(); err == nil {
		t.Fatal("cross-tenant evidence must be rejected")
	}
	status := AIRunStatusRecord{SchemaVersion: AIContractVersion, RunID: "run-1", RequestID: "req-1", TenantID: "tenant-a", Status: "unknown"}
	if err := status.Validate(); err == nil {
		t.Fatal("unknown run status must be rejected")
	}
	result := AIDiagnosisResult{SchemaVersion: AIContractVersion, RequestID: "req-1", TenantID: "tenant-a", Outcome: AIDiagnosisOutcomeInsufficientEvidence, Confidence: AIConfidence{Level: AIConfidenceUnknown}, ProviderError: &AIProviderError{Class: "provider_secret_leak"}}
	if err := result.Validate(); err == nil {
		t.Fatal("unknown provider error class must be rejected")
	}
}
