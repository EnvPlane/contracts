package domain

type PolicyEffect string

const (
	PolicyAllow PolicyEffect = "allow"
	PolicyDeny  PolicyEffect = "deny"
	PolicyWarn  PolicyEffect = "warn"
)

type PolicyInput struct {
	TenantID       string `json:"tenantId"`
	Action         string `json:"action"`
	QuotaAllowed   bool   `json:"quotaAllowed"`
	ProjectReady   bool   `json:"projectReady"`
	TTLAllowed     bool   `json:"ttlAllowed"`
	PinAllowed     bool   `json:"pinAllowed"`
	RepairApproved bool   `json:"repairApproved"`
}

type PolicyDecision struct {
	Effect        PolicyEffect `json:"effect"`
	Allowed       bool         `json:"allowed"`
	RuleIDs       []string     `json:"ruleIds"`
	Explanation   string       `json:"explanation"`
	BundleVersion string       `json:"bundleVersion"`
}
