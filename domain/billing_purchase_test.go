package domain

import "testing"

func TestBillingEvidenceAndIssuanceTransitions(t *testing.T) {
	states := []BillingPaymentState{BillingAwaitingPayment, BillingPaymentFailed, BillingPaymentPaid, BillingPaymentNotCharged, "unknown"}
	for _, from := range states {
		for _, to := range states {
			expected := from == BillingAwaitingPayment && (to == BillingAwaitingPayment || to == BillingPaymentFailed || to == BillingPaymentPaid) || from == BillingPaymentFailed && (to == BillingPaymentFailed || to == BillingPaymentPaid) || from == BillingPaymentPaid && to == BillingPaymentPaid || from == BillingPaymentNotCharged && to == BillingPaymentNotCharged
			if ValidBillingPaymentTransition(from, to) != expected {
				t.Fatalf("payment transition %s to %s", from, to)
			}
		}
	}
	for _, from := range []BillingIssuanceState{BillingIssuancePending, BillingIssuanceIssued, "unknown"} {
		for _, to := range []BillingIssuanceState{BillingIssuancePending, BillingIssuanceIssued, "unknown"} {
			expected := from == BillingIssuancePending && (to == BillingIssuancePending || to == BillingIssuanceIssued) || from == BillingIssuanceIssued && to == BillingIssuanceIssued
			if ValidBillingIssuanceTransition(from, to) != expected {
				t.Fatalf("issuance transition %s to %s", from, to)
			}
		}
	}
}
