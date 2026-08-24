package domain

import "testing"

func TestTenantAIPolicyIsDisabledAndRedactionCannotBeRelaxed(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	if policy.Mode != AIPolicyDisabled || policy.MaxContextClassification != "safe_metadata" {
		t.Fatal("default AI policy is not disabled and safe")
	}
	policy.MaxContextClassification = "raw_payload"
	if err := policy.Validate(); err == nil {
		t.Fatal("tenant policy relaxed the platform context classification")
	}
}

func TestTenantAIPolicyRejectsUnknownPurpose(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	policy.Purposes["arbitrary_write"] = true
	if err := policy.Validate(); err == nil {
		t.Fatal("unknown AI purpose was accepted")
	}
}

func TestOfflineTenantPolicyIsAValidNoEgressMode(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	policy.Mode = AIPolicyOffline
	policy.Purposes["diagnosis"] = true
	if err := policy.Validate(); err != nil {
		t.Fatalf("offline policy rejected: %v", err)
	}
}
