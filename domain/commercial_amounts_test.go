package domain

import (
	"encoding/json"
	"testing"
)

func TestInclusiveCommercialAmountsAndLegacyCompatibility(t *testing.T) {
	legacy := CommercialSnapshot{Currency: "EUR", AmountMinor: 1000, BillingInterval: "month", TaxMode: "exclusive"}
	raw, _ := json.Marshal(legacy)
	if string(raw) != `{"currency":"EUR","amountMinor":1000,"billingInterval":"month","taxMode":"exclusive"}` || legacy.ValidateAmounts() != nil {
		t.Fatal("legacy snapshot changed", string(raw))
	}
	for _, gross := range []int64{9900, 29900} {
		a := CommercialAmounts{TaxPresentationVersion: InclusiveTaxPresentationVersion, GrantKind: "paid", ListPriceMinor: gross, SubtotalMinor: gross - 1581, TaxMinor: 1581, TotalMinor: gross, ChargedTotalMinor: gross}
		s := CommercialSnapshot{Currency: "EUR", AmountMinor: gross, BillingInterval: "month", TaxMode: "inclusive", Amounts: &a}
		if s.ValidateAmounts() != nil {
			t.Fatal("inclusive gross rejected", s)
		}
		for _, mutate := range []func(*CommercialAmounts){
			func(a *CommercialAmounts) { a.TaxPresentationVersion = "unknown" }, func(a *CommercialAmounts) { a.SubtotalMinor++ }, func(a *CommercialAmounts) { a.DiscountMinor = 1 }, func(a *CommercialAmounts) { a.TaxMinor = -1 }, func(a *CommercialAmounts) { a.ChargedTotalMinor-- }, func(a *CommercialAmounts) { a.GrantKind = "trial" }, func(a *CommercialAmounts) { a.ListPriceMinor++ },
		} {
			copy := a
			mutate(&copy)
			if copy.Validate() == nil {
				t.Fatal("unsupported/drifting amounts accepted", copy)
			}
		}
		s.AmountMinor--
		if s.ValidateAmounts() == nil {
			t.Fatal("license amount replaced actual charged total")
		}
	}
	grace := CommercialAmounts{TaxPresentationVersion: InclusiveTaxPresentationVersion, GrantKind: "grace", ListPriceMinor: 9900}
	if grace.Validate() != nil {
		t.Fatal("zero-charge grace rejected")
	}
	grace.TotalMinor = 9900
	grace.SubtotalMinor = 9900
	grace.ChargedTotalMinor = 9900
	if grace.Validate() == nil {
		t.Fatal("grace claims new payment")
	}
}

func TestMerchantProfileBlocksUnknownApplicableTaxPolicy(t *testing.T) {
	if (MerchantInvoiceProfile{}).ValidateLiveLaunch() == nil {
		t.Fatal("unknown tax policy permits live")
	}
	p := MerchantInvoiceProfile{Version: "merchant-invoice-v1", ProviderMode: "live", ProviderAccountRef: "acct_fixture", LegalName: "Fixture Seller", AddressLines: []string{"Fixture Address"}, MerchantDetailsApprovalRef: "fixture-merchant-review", TaxPolicyApprovalRef: "fixture-accountant-review", TaxPresentationVersion: InclusiveTaxPresentationVersion, ReviewStatus: "owner-accountant-approved", CustomerDataPolicy: "ownership-and-financial-only"}
	if p.ValidateLiveLaunch() != nil {
		t.Fatal("reviewed fixture profile rejected")
	}
	for _, mutate := range []func(*MerchantInvoiceProfile){func(p *MerchantInvoiceProfile) { p.TaxPolicyApprovalRef = "" }, func(p *MerchantInvoiceProfile) { p.MerchantDetailsApprovalRef = "" }, func(p *MerchantInvoiceProfile) { p.ProviderMode = "test" }, func(p *MerchantInvoiceProfile) { p.ReviewStatus = "pending" }, func(p *MerchantInvoiceProfile) { p.LegalName = "" }, func(p *MerchantInvoiceProfile) { p.AddressLines = nil }, func(p *MerchantInvoiceProfile) { p.TaxPresentationVersion = "exclusive" }} {
		copy := p
		mutate(&copy)
		if copy.ValidateLiveLaunch() == nil {
			t.Fatal("unapproved live profile accepted", copy)
		}
	}
}
