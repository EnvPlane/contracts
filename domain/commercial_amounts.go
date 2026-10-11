package domain

import "errors"

const InclusiveTaxPresentationVersion = "inclusive-gross-v1"

// CommercialAmounts is signed financial display evidence, never an entitlement
// input or a decision about a jurisdiction/rate. Subtotal is net of inclusive tax.
type CommercialAmounts struct {
	TaxPresentationVersion string `json:"taxPresentationVersion"`
	GrantKind              string `json:"grantKind"`
	ListPriceMinor         int64  `json:"listPriceMinor"`
	SubtotalMinor          int64  `json:"subtotalMinor"`
	DiscountMinor          int64  `json:"discountMinor"`
	TaxMinor               int64  `json:"taxMinor"`
	TotalMinor             int64  `json:"totalMinor"`
	ChargedTotalMinor      int64  `json:"chargedTotalMinor"`
}

func (a CommercialAmounts) Validate() error {
	reject := func() error { return errors.New("invalid normalized commercial amounts") }
	if a.TaxPresentationVersion != InclusiveTaxPresentationVersion || a.ListPriceMinor <= 0 || a.SubtotalMinor < 0 || a.DiscountMinor != 0 || a.TaxMinor < 0 || a.TotalMinor < 0 || a.ChargedTotalMinor < 0 || a.TaxMinor > a.TotalMinor || a.SubtotalMinor != a.TotalMinor-a.TaxMinor || a.ChargedTotalMinor != a.TotalMinor {
		return reject()
	}
	switch a.GrantKind {
	case "paid":
		if a.TotalMinor != a.ListPriceMinor {
			return reject()
		}
	case "grace":
		if a.TotalMinor != 0 {
			return reject()
		}
	default:
		return reject()
	}
	return nil
}

// ValidateAmounts preserves legacy snapshots without rewriting their semantics.
// When a normalized breakdown is present, AmountMinor is actual paid total (zero
// for grace) and must not silently substitute today's catalog/list price.
func (s CommercialSnapshot) ValidateAmounts() error {
	if s.Amounts == nil {
		return nil
	}
	if s.Amounts.Validate() != nil || s.TaxMode != "inclusive" || s.BillingInterval != "month" || s.Currency != "EUR" || s.AmountMinor != s.Amounts.ChargedTotalMinor {
		return errors.New("commercial snapshot and normalized amounts disagree")
	}
	return nil
}
