package domain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"testing"
	"time"
)

func TestActivationCodeRejectsTamperBindingReplayUnknownKeyAndExpiry(t *testing.T) {
	seed := sha256.Sum256([]byte("activation-vector-v1"))
	private := ed25519.NewKeyFromSeed(seed[:])
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	grant := ActivationGrant{SchemaVersion: ActivationCodeSchemaVersion, InstallationID: "install-a", TenantID: "tenant-a", SKU: "pro", PlanID: "pro", PlanVersion: "2026.1", Features: map[string]bool{"gitops": true}, Limits: map[string]int64{"projects": 5}, IssuedAt: now, NotBefore: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), LicenseID: "lic-1", Nonce: "nonce-1", Commercial: CommercialSnapshot{Currency: "EUR", AmountMinor: 1000, BillingInterval: "month", TaxMode: "exclusive"}}
	code, err := SignActivationCode(grant, "issuer-1", private)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]ed25519.PublicKey{"issuer-1": private.Public().(ed25519.PublicKey)}
	if _, err := VerifyActivationCode(code, keys, ActivationCodeBinding{"install-a", "tenant-a"}, now, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyActivationCode(code+"x", keys, ActivationCodeBinding{"install-a", "tenant-a"}, now, nil); err == nil {
		t.Fatal("tamper accepted")
	}
	if _, err := VerifyActivationCode(code, keys, ActivationCodeBinding{"install-b", "tenant-a"}, now, nil); err == nil {
		t.Fatal("binding accepted")
	}
	if _, err := VerifyActivationCode(code, keys, ActivationCodeBinding{"install-a", "tenant-a"}, now, func(string) bool { return true }); err == nil {
		t.Fatal("replay accepted")
	}
	if _, err := VerifyActivationCode(code, map[string]ed25519.PublicKey{}, ActivationCodeBinding{"install-a", "tenant-a"}, now, nil); err == nil {
		t.Fatal("unknown key accepted")
	}
	if _, err := VerifyActivationCode(code, keys, ActivationCodeBinding{"install-a", "tenant-a"}, now.Add(2*time.Hour), nil); err == nil {
		t.Fatal("expiry accepted")
	}
	if allowed, limit := grant.Authorized("gitops", "projects"); !allowed || limit != 5 {
		t.Fatal("authorization contract changed")
	}
}
