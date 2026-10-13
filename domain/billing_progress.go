package domain

import "time"

// BillingProgress is seller-owned evidence, not a second entitlement source.
// Local signed activation is deliberately a separate read.
type BillingProgress struct {
	TenantID     string                  `json:"tenantId"`
	ProviderMode string                  `json:"providerMode"`
	Records      []BillingProgressRecord `json:"records"`
}
type BillingProgressRecord struct {
	AttemptID         string             `json:"attemptId"`
	PlanID            string             `json:"planId"`
	PlanVersion       string             `json:"planVersion"`
	SubscriptionState string             `json:"subscriptionState"`
	PaymentState      string             `json:"paymentState"`
	IssuanceState     string             `json:"issuanceState"`
	State             string             `json:"state"`
	PurchaseID        string             `json:"purchaseId,omitempty"`
	CurrentLicenseID  string             `json:"currentLicenseId,omitempty"`
	LicenseID         string             `json:"licenseId,omitempty"`
	ExpiresAt         *time.Time         `json:"expiresAt,omitempty"`
	GraceEndsAt       *time.Time         `json:"graceEndsAt,omitempty"`
	Amounts           *CommercialAmounts `json:"amounts,omitempty"`
	Currency          string             `json:"currency,omitempty"`
	InvoiceURL        string             `json:"invoiceUrl,omitempty"`
	RetryAllowed      bool               `json:"retryAllowed"`
}
type BillingIssuanceRetry struct {
	RequestID  string `json:"requestId"`
	PurchaseID string `json:"purchaseId"`
}

// Validate rejects ambiguous or foreign status rather than guessing payment.
func (p BillingProgress) Validate(tenant string) bool { return p.ValidateMode(tenant, "test") }

// ValidateMode rejects provider-mode crossover against configured server state.
func (p BillingProgress) ValidateMode(tenant, mode string) bool {
	if p.TenantID != tenant || tenant == "" || ((mode != "test" && mode != "live") || p.ProviderMode != mode) || p.Records == nil || len(p.Records) > 25 {
		return false
	}
	seen := map[string]bool{}
	for _, r := range p.Records {
		if r.AttemptID == "" || seen[r.AttemptID] || r.PlanID == "" || r.PlanVersion == "" {
			return false
		}
		seen[r.AttemptID] = true
		switch r.State {
		case "checkout_pending", "payment_processing", "paid/issuance_pending", "activation_required", "past_due", "canceled", "failed":
		default:
			return false
		}
		switch r.PaymentState {
		case "awaiting_payment", "payment_failed", "paid", "not_charged":
		default:
			return false
		}
		switch r.IssuanceState {
		case "not_started", "issuance_pending", "issued":
		default:
			return false
		}
		if r.Amounts != nil && (r.Amounts.Validate() != nil || r.Currency != "EUR") {
			return false
		}
		if r.State == "paid/issuance_pending" && (r.PaymentState != "paid" || r.IssuanceState != "issuance_pending" || r.PurchaseID == "" || r.LicenseID != "") {
			return false
		}
		if r.State == "activation_required" && (r.PaymentState != "paid" || r.IssuanceState != "issued" || r.LicenseID == "") {
			return false
		}
		if r.RetryAllowed && r.State != "paid/issuance_pending" {
			return false
		}
	}
	return true
}

func (r BillingIssuanceRetry) Valid() bool {
	valid := func(s string, min, max int) bool {
		if len(s) < min || len(s) > max {
			return false
		}
		for _, c := range s {
			if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '.' && c != '_' && c != ':' && c != '-' {
				return false
			}
		}
		return true
	}
	return valid(r.RequestID, 8, 128) && valid(r.PurchaseID, 1, 128)
}
