package domain

import (
	"fmt"
	"sort"
	"strings"
)

const PolicyBundleSchemaVersion = "1"

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
	// Deprecated: budget enforcement is evaluated by the server budget
	// dependency, not by the policy engine.
	BudgetChecked bool `json:"budgetChecked,omitempty"`
	// Deprecated: retained for wire compatibility; ignored by the policy engine.
	BudgetAllowed bool `json:"budgetAllowed,omitempty"`
	// Deprecated: retained for wire compatibility; ignored by the policy engine.
	BudgetWarning bool `json:"budgetWarning,omitempty"`
}

type PolicyDecision struct {
	Effect        PolicyEffect `json:"effect"`
	Allowed       bool         `json:"allowed"`
	RuleIDs       []string     `json:"ruleIds"`
	Explanation   string       `json:"explanation"`
	BundleVersion string       `json:"bundleVersion"`
}

type PolicyRule struct {
	ID          string       `json:"id"`
	Action      string       `json:"action"`
	Check       string       `json:"check"`
	Effect      PolicyEffect `json:"effect"`
	Explanation string       `json:"explanation"`
}

type PolicyBundle struct {
	SchemaVersion string       `json:"schemaVersion"`
	BundleVersion string       `json:"bundleVersion"`
	Rules         []PolicyRule `json:"rules"`
}

func BuiltinPolicyBundle(version string) PolicyBundle {
	if strings.TrimSpace(version) == "" {
		version = "1.0.0"
	}
	return PolicyBundle{SchemaVersion: PolicyBundleSchemaVersion, BundleVersion: version, Rules: []PolicyRule{
		{ID: "POL-ENV-001", Action: "create_environment", Check: "project_ready", Effect: PolicyDeny, Explanation: "project is not deployment-ready"},
		{ID: "POL-QUOTA-001", Action: "create_environment", Check: "quota_allowed", Effect: PolicyDeny, Explanation: "entitlement quota is exhausted"},
		{ID: "POL-TTL-001", Action: "create_environment", Check: "ttl_allowed", Effect: PolicyDeny, Explanation: "requested TTL exceeds entitlement"},
		{ID: "POL-PIN-001", Action: "pin_environment", Check: "pin_allowed", Effect: PolicyDeny, Explanation: "requested pin exceeds entitlement"},
		{ID: "POL-CLUSTER-001", Action: "onboard_cluster", Check: "quota_allowed", Effect: PolicyDeny, Explanation: "cluster entitlement is exhausted"},
		{ID: "POL-REPAIR-001", Action: "privileged_repair", Check: "repair_approved", Effect: PolicyDeny, Explanation: "privileged repair requires approval"},
	}}
}

func (b PolicyBundle) Validate() error {
	if b.SchemaVersion != PolicyBundleSchemaVersion || strings.TrimSpace(b.BundleVersion) == "" {
		return fmt.Errorf("invalid policy bundle version")
	}
	seen := map[string]struct{}{}
	for _, rule := range b.Rules {
		if strings.TrimSpace(rule.ID) == "" || strings.TrimSpace(rule.Action) == "" || strings.TrimSpace(rule.Check) == "" {
			return fmt.Errorf("policy rule id, action and check are required")
		}
		if _, ok := seen[rule.ID]; ok {
			return fmt.Errorf("duplicate policy rule %q", rule.ID)
		}
		seen[rule.ID] = struct{}{}
		switch rule.Check {
		case "project_ready", "quota_allowed", "ttl_allowed", "pin_allowed", "repair_approved":
		default:
			return fmt.Errorf("unknown policy check %q", rule.Check)
		}
		if rule.Effect != PolicyAllow && rule.Effect != PolicyDeny && rule.Effect != PolicyWarn {
			return fmt.Errorf("invalid policy effect for %q", rule.ID)
		}
	}
	return nil
}

func (b PolicyBundle) Deterministic() PolicyBundle {
	copyBundle := PolicyBundle{SchemaVersion: b.SchemaVersion, BundleVersion: b.BundleVersion, Rules: append([]PolicyRule(nil), b.Rules...)}
	sort.SliceStable(copyBundle.Rules, func(i, j int) bool { return copyBundle.Rules[i].ID < copyBundle.Rules[j].ID })
	return copyBundle
}
