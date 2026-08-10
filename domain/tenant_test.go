package domain

import "testing"

func TestNormalizeTenantIDUsesStableDefaultForLegacyRecords(t *testing.T) {
	if got := NormalizeTenantID(""); got != DefaultTenantID {
		t.Fatalf("empty tenant id = %q, want %q", got, DefaultTenantID)
	}
	if got := NormalizeTenantID("tenant-a"); got != "tenant-a" {
		t.Fatalf("explicit tenant id = %q", got)
	}
}
