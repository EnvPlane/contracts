package domain

import "fmt"

const (
	QuotaFeatureDisabled = "quota_feature_disabled"
	QuotaExhausted       = "quota_exhausted"
)

// QuotaDecisionInput is pure input; evaluating it never mutates usage or entitlements.
type QuotaDecisionInput struct {
	TenantID       string
	Feature        string
	Limit          string
	Current        int64
	RequestedDelta int64
	Snapshot       EntitlementSnapshot
	RequestID      string
	UpgradeURL     string
}

type QuotaError struct {
	Code       string `json:"code"`
	Feature    string `json:"feature,omitempty"`
	Limit      string `json:"limit,omitempty"`
	Current    int64  `json:"current,omitempty"`
	Requested  int64  `json:"requested,omitempty"`
	Plan       string `json:"plan,omitempty"`
	UpgradeURL string `json:"upgrade_url,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e QuotaError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Limit) }

type QuotaDecision struct {
	Allowed    bool        `json:"allowed"`
	HTTPStatus int         `json:"-"`
	Error      *QuotaError `json:"error,omitempty"`
}
