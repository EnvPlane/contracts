package installationbilling

import "strings"

// These additions are negotiated by exact operations; old v1 JSON/frames are unchanged.
func ActionOperation(op string) bool {
	return op == "checkout" || op == "portal" || op == "status" || op == "paid.status" || op == "redeem"
}

type PaidStatusBody struct {
	AttemptID string `json:"attemptId"`
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
}
type RedeemBody struct {
	AttemptID string `json:"attemptId"`
	Nonce     string `json:"nonce"`
}
type PaidStatus struct {
	Records []PaidRecord `json:"records"`
	Cursor  string       `json:"cursor"`
}

// Times are canonical Unix seconds; 0 means no established deadline. This is a
// bounded observation, not permission to issue or a substitute for commit guards.
type PaidRecord struct {
	AttemptID         string `json:"attemptId"`
	PlanID            string `json:"planId"`
	PlanVersion       string `json:"planVersion"`
	SubscriptionState string `json:"subscriptionState"`
	PaymentState      string `json:"paymentState"`
	IssuanceState     string `json:"issuanceState"`
	State             string `json:"state"`
	PurchaseID        string `json:"purchaseId"`
	LicenseID         string `json:"licenseId"`
	CurrentLicenseID  string `json:"currentLicenseId"`
	ExpiresAt         string `json:"expiresAt"`
	GraceEndsAt       string `json:"graceEndsAt"`
}

func ValidatePaidStatus(p PaidStatus) error {
	if p.Records == nil || len(p.Records) > 25 || (p.Cursor != "" && !validText(p.Cursor, 128)) {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for _, r := range p.Records {
		if !validText(r.AttemptID, 128) || seen[r.AttemptID] || !validText(r.PlanID, 128) || !validText(r.PlanVersion, 128) || !validText(r.SubscriptionState, 128) {
			return ErrInvalid
		}
		seen[r.AttemptID] = true
		for _, s := range []string{r.PurchaseID, r.LicenseID, r.CurrentLicenseID} {
			if s != "" && !validText(s, 128) {
				return ErrInvalid
			}
		}
		for _, s := range []string{r.ExpiresAt, r.GraceEndsAt} {
			if v, e := decimal(s, true); e != nil || v > 253402300799 {
				return ErrInvalid
			}
		}
		if !oneOf(r.PaymentState, "awaiting_payment", "payment_failed", "paid", "not_charged") || !oneOf(r.IssuanceState, "not_started", "issuance_pending", "issued") || !oneOf(r.State, "checkout_pending", "payment_processing", "paid/issuance_pending", "activation_required", "past_due", "canceled", "failed") {
			return ErrInvalid
		}
		if r.State == "activation_required" && (r.PaymentState != "paid" || r.IssuanceState != "issued" || r.LicenseID == "" || r.CurrentLicenseID != r.LicenseID || r.PurchaseID == "" || r.ExpiresAt == "0") {
			return ErrInvalid
		}
		if r.State == "paid/issuance_pending" && (r.PaymentState != "paid" || r.IssuanceState != "issuance_pending" || r.PurchaseID == "" || r.LicenseID != "") {
			return ErrInvalid
		}
	}
	return nil
}
func oneOf(s string, ss ...string) bool {
	for _, x := range ss {
		if s == x {
			return true
		}
	}
	return false
}
func validatePaidResponse(r Response) error {
	if r.Operation == "paid.status" || r.Operation == "redeem" {
		if r.PaidStatus == nil || ValidatePaidStatus(*r.PaidStatus) != nil {
			return ErrInvalid
		}
	} else if r.PaidStatus != nil {
		return ErrInvalid
	}
	if r.Operation == "redeem" {
		if len(r.PaidStatus.Records) != 1 || r.PaidStatus.Cursor != "" || r.PaidStatus.Records[0].State != "activation_required" || len(r.ActivationCode) > 8192 || !strings.HasPrefix(r.ActivationCode, "v1.") || strings.ContainsAny(r.ActivationCode, " \t\r\n") {
			return ErrInvalid
		}
	} else if r.ActivationCode != "" {
		return ErrInvalid
	}
	return nil
}
