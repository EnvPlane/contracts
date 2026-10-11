package domain

// Payment evidence and issuance progress are independent: an unavailable issuer
// never changes a confirmed paid receipt back into an unpaid purchase.
type BillingPaymentState string
type BillingIssuanceState string

const (
	BillingAwaitingPayment   BillingPaymentState  = "awaiting_payment"
	BillingPaymentFailed     BillingPaymentState  = "payment_failed"
	BillingPaymentPaid       BillingPaymentState  = "paid"
	BillingPaymentNotCharged BillingPaymentState  = "not_charged" // existing approved renewal grace only
	BillingIssuancePending   BillingIssuanceState = "issuance_pending"
	BillingIssuanceIssued    BillingIssuanceState = "issued"
)

func ValidBillingPaymentTransition(from, to BillingPaymentState) bool {
	switch from {
	case BillingAwaitingPayment:
		return to == BillingAwaitingPayment || to == BillingPaymentFailed || to == BillingPaymentPaid
	case BillingPaymentFailed:
		return to == BillingPaymentFailed || to == BillingPaymentPaid
	case BillingPaymentPaid:
		return to == BillingPaymentPaid
	case BillingPaymentNotCharged:
		return to == BillingPaymentNotCharged
	}
	return false
}
func ValidBillingIssuanceTransition(from, to BillingIssuanceState) bool {
	return from == BillingIssuancePending && (to == BillingIssuancePending || to == BillingIssuanceIssued) || from == BillingIssuanceIssued && to == BillingIssuanceIssued
}
