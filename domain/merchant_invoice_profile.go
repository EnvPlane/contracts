package domain

import (
	"errors"
	"strings"
)

// MerchantInvoiceProfile is seller-controlled operator configuration. Approval
// references identify externally reviewed records; no caller metadata grants approval.
type MerchantInvoiceProfile struct {
	Version                    string   `json:"version"`
	ProviderMode               string   `json:"providerMode"`
	ProviderAccountRef         string   `json:"providerAccountRef"`
	LegalName                  string   `json:"legalName"`
	AddressLines               []string `json:"addressLines"`
	MerchantDetailsApprovalRef string   `json:"merchantDetailsApprovalRef"`
	TaxPolicyApprovalRef       string   `json:"taxPolicyApprovalRef"`
	TaxPresentationVersion     string   `json:"taxPresentationVersion"`
	ReviewStatus               string   `json:"reviewStatus"`
	CustomerDataPolicy         string   `json:"customerDataPolicy"`
}

// ValidateLiveLaunch is a prerequisite for future live enablement, not an
// instruction to enable it. Current issuer remains test-only regardless of this profile.
func (p MerchantInvoiceProfile) ValidateLiveLaunch() error {
	if p.Version != "merchant-invoice-v1" || p.ProviderMode != "live" || !strings.HasPrefix(p.ProviderAccountRef, "acct_") || len(p.ProviderAccountRef) <= 5 || strings.TrimSpace(p.LegalName) == "" || len(p.AddressLines) == 0 || len(p.AddressLines) > 6 || strings.TrimSpace(p.MerchantDetailsApprovalRef) == "" || strings.TrimSpace(p.TaxPolicyApprovalRef) == "" || p.TaxPresentationVersion != InclusiveTaxPresentationVersion || p.ReviewStatus != "owner-accountant-approved" || p.CustomerDataPolicy != "ownership-and-financial-only" {
		return errors.New("live launch requires an approved merchant invoice profile and applicable tax policy")
	}
	for _, line := range p.AddressLines {
		if strings.TrimSpace(line) == "" {
			return errors.New("merchant invoice address incomplete")
		}
	}
	return nil
}
