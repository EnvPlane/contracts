package domain

import (
	"testing"
	"time"
)

func TestAISecurityFindingRejectsProtectedPayload(t *testing.T) {
	now := time.Now().UTC()
	finding := AISecurityFinding{ID: "finding-1", Kind: AISecurityAuditAnomaly, Severity: AISecurityMedium, TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "env-a", Source: AIEvidenceReference{TenantID: "tenant-a", SourceType: "audit", SourceID: "event-1"}, Subject: "authorization: Bearer secret"}
	if err := finding.Validate("tenant-a", "project-a", "env-a", now); err == nil {
		t.Fatal("protected payload must be rejected")
	}
}

func TestAISecurityExceptionRequiresAuditedUnexpiredApprover(t *testing.T) {
	now := time.Now().UTC()
	exception := AISecurityException{ID: "exception-1", TenantID: "tenant-a", ProjectID: "project-a", FindingID: "finding-1", Reason: "approved maintenance", ApproverID: "operator-1", ExpiresAt: now.Add(time.Hour), Audited: false}
	if err := exception.Validate("tenant-a", "project-a", "finding-1", now); err == nil {
		t.Fatal("unaudited exception must be rejected")
	}
	exception.Audited = true
	if err := exception.Validate("tenant-a", "project-a", "finding-1", now); err != nil {
		t.Fatal(err)
	}
	exception.ExpiresAt = now.Add(-time.Second)
	if err := exception.Validate("tenant-a", "project-a", "finding-1", now); err == nil {
		t.Fatal("expired exception must be rejected")
	}
}
