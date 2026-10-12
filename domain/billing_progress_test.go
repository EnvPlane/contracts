package domain

import "testing"

func TestBillingProgressRejectsForeignAndUnpaidIssuance(t *testing.T) {
	p := BillingProgress{TenantID: "t", ProviderMode: "test", Records: []BillingProgressRecord{{AttemptID: "attempt-1", PlanID: "team", PlanVersion: "2.0.0", State: "activation_required", PaymentState: "paid", IssuanceState: "issued", LicenseID: "license-1"}}}
	if !p.Validate("t") || p.Validate("foreign") {
		t.Fatal("scope validation")
	}
	p.Records[0].PaymentState = "awaiting_payment"
	if p.Validate("t") {
		t.Fatal("unpaid issuance trusted")
	}
	p.Records[0].PaymentState = "paid"
	p.Records[0].RetryAllowed = true
	if p.Validate("t") {
		t.Fatal("completed retry enabled")
	}
	p.Records[0].RetryAllowed = false
	p.ProviderMode = "live"
	if p.Validate("t") {
		t.Fatal("mode crossover")
	}
}
