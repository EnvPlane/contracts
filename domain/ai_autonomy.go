package domain

import (
	"fmt"
	"strings"
)

const AIAutonomyContractVersion = "1"

type AIAutonomyLevel string

const (
	AIAutonomyObserve           AIAutonomyLevel = "observe"
	AIAutonomyRecommend         AIAutonomyLevel = "recommend"
	AIAutonomyApprovalRequired  AIAutonomyLevel = "approval_required"
	AIAutonomyAutonomousBounded AIAutonomyLevel = "autonomous_bounded"
)

type AIRiskClass string

const (
	AIRiskLow      AIRiskClass = "low"
	AIRiskMedium   AIRiskClass = "medium"
	AIRiskHigh     AIRiskClass = "high"
	AIRiskCritical AIRiskClass = "critical"
)

type AIActionClass string

const (
	AIActionReadOnly      AIActionClass = "read_only"
	AIActionProposal      AIActionClass = "proposal"
	AIActionApprovedWrite AIActionClass = "approved_write"
	AIActionForbidden     AIActionClass = "forbidden"
)

type AICapability struct {
	SchemaVersion   string          `json:"schemaVersion"`
	Purpose         string          `json:"purpose"`
	Action          string          `json:"action"`
	DefaultAutonomy AIAutonomyLevel `json:"defaultAutonomy"`
	Risk            AIRiskClass     `json:"risk"`
	ActionClass     AIActionClass   `json:"actionClass"`
	RequiredFeature string          `json:"requiredFeature"`
	RequiredRoles   []string        `json:"requiredRoles"`
	AuditEvents     []string        `json:"auditEvents"`
}

type AIEffectiveCapability struct {
	AICapability
	Known             bool            `json:"known"`
	Enabled           bool            `json:"enabled"`
	EffectiveAutonomy AIAutonomyLevel `json:"effectiveAutonomy"`
	Reason            string          `json:"reason"`
}

type AICapabilityMatrix struct {
	SchemaVersion    string                  `json:"schemaVersion"`
	TenantID         string                  `json:"tenantId"`
	MaxAutonomy      AIAutonomyLevel         `json:"maxAutonomy"`
	ProviderMode     AIPolicyMode            `json:"providerMode"`
	ActivationReason string                  `json:"activationReason"`
	Quotas           map[string]int64        `json:"quotas"`
	Capabilities     []AIEffectiveCapability `json:"capabilities"`
}

var aiCapabilities = []AICapability{
	{Purpose: "diagnosis", Action: "environment.explain_failure", DefaultAutonomy: AIAutonomyRecommend, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.diagnosis", RequiredRoles: []string{"projects.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "bootstrap_troubleshooting", Action: "bootstrap.explain_blocked_step", DefaultAutonomy: AIAutonomyRecommend, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.bootstrap", RequiredRoles: []string{"projects.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "bootstrap.scan", Action: "bootstrap.scan_repository", DefaultAutonomy: AIAutonomyObserve, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.bootstrap", RequiredRoles: []string{"projects.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "configuration_proposal", Action: "configuration.propose", DefaultAutonomy: AIAutonomyApprovalRequired, Risk: AIRiskMedium, ActionClass: AIActionProposal, RequiredFeature: "ai.configuration", RequiredRoles: []string{"projects.write"}, AuditEvents: []string{"ai.proposal.created", "ai.proposal.confirmed"}},
	{Purpose: "environment_creation", Action: "environment.create", DefaultAutonomy: AIAutonomyApprovalRequired, Risk: AIRiskHigh, ActionClass: AIActionApprovedWrite, RequiredFeature: "ai.environment_create", RequiredRoles: []string{"environments.write"}, AuditEvents: []string{"ai.proposal.created", "ai.approval.created", "ai.action.executed"}},
	{Purpose: "finops_explanation", Action: "finops.explain", DefaultAutonomy: AIAutonomyRecommend, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.finops", RequiredRoles: []string{"projects.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "gitops.diagnosis", Action: "gitops.explain_or_reconcile", DefaultAutonomy: AIAutonomyRecommend, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.gitops", RequiredRoles: []string{"projects.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "kubernetes.diagnosis", Action: "kubernetes.explain_or_reconcile", DefaultAutonomy: AIAutonomyRecommend, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.kubernetes", RequiredRoles: []string{"environments.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "scm.diagnosis", Action: "scm.explain_or_propose", DefaultAutonomy: AIAutonomyRecommend, Risk: AIRiskLow, ActionClass: AIActionReadOnly, RequiredFeature: "ai.scm", RequiredRoles: []string{"environments.read"}, AuditEvents: []string{"ai.run.created", "ai.run.completed"}},
	{Purpose: "approved_actions", Action: "environment.retry_or_refresh", DefaultAutonomy: AIAutonomyApprovalRequired, Risk: AIRiskHigh, ActionClass: AIActionApprovedWrite, RequiredFeature: "ai.approved_actions", RequiredRoles: []string{"environments.write"}, AuditEvents: []string{"ai.proposal.confirmed", "ai.approval.created", "ai.action.executed"}},
	{Purpose: "approved_actions", Action: "environment.delete", DefaultAutonomy: AIAutonomyObserve, Risk: AIRiskCritical, ActionClass: AIActionForbidden, RequiredFeature: "never", RequiredRoles: []string{"none"}, AuditEvents: []string{"ai.action.denied"}},
}

func AICapabilityCatalog() []AICapability {
	result := make([]AICapability, len(aiCapabilities))
	copy(result, aiCapabilities)
	for i := range result {
		result[i].SchemaVersion = AIAutonomyContractVersion
		result[i].RequiredRoles = append([]string(nil), result[i].RequiredRoles...)
		result[i].AuditEvents = append([]string(nil), result[i].AuditEvents...)
	}
	return result
}

func (p TenantAIPolicy) EffectiveMaxAutonomy() AIAutonomyLevel {
	if p.MaxAutonomy == "" {
		return AIAutonomyObserve
	}
	return p.MaxAutonomy
}

func autonomyRank(level AIAutonomyLevel) int {
	switch level {
	case AIAutonomyObserve:
		return 0
	case AIAutonomyRecommend:
		return 1
	case AIAutonomyApprovalRequired:
		return 2
	case AIAutonomyAutonomousBounded:
		return 3
	default:
		return -1
	}
}

func (c AICapability) Validate() error {
	if c.SchemaVersion != AIAutonomyContractVersion || strings.TrimSpace(c.Purpose) == "" || strings.TrimSpace(c.Action) == "" {
		return fmt.Errorf("invalid AI capability identity or schema")
	}
	if autonomyRank(c.DefaultAutonomy) < 0 {
		return fmt.Errorf("invalid AI capability autonomy or risk")
	}
	switch c.Risk {
	case AIRiskLow, AIRiskMedium, AIRiskHigh, AIRiskCritical:
	default:
		return fmt.Errorf("unsupported AI risk class %q", c.Risk)
	}
	switch c.ActionClass {
	case AIActionReadOnly, AIActionProposal, AIActionApprovedWrite, AIActionForbidden:
	default:
		return fmt.Errorf("unsupported AI action class %q", c.ActionClass)
	}
	return nil
}

func ResolveAICapability(policy TenantAIPolicy, purpose, action string) AIEffectiveCapability {
	for _, capability := range aiCapabilities {
		if capability.Purpose != purpose || capability.Action != action {
			continue
		}
		capability.SchemaVersion = AIAutonomyContractVersion
		result := AIEffectiveCapability{AICapability: capability, Known: true, EffectiveAutonomy: capability.DefaultAutonomy}
		if capability.ActionClass == AIActionForbidden {
			result.Reason = "action_forbidden_by_platform_policy"
			return result
		}
		if policy.Mode == AIPolicyDisabled {
			result.Enabled, result.EffectiveAutonomy, result.Reason = false, AIAutonomyObserve, "tenant_ai_disabled"
			return result
		}
		if !policy.PurposeEnabled(purpose) {
			result.Enabled, result.EffectiveAutonomy, result.Reason = false, AIAutonomyObserve, "purpose_disabled_by_tenant_policy"
			return result
		}
		max := policy.EffectiveMaxAutonomy()
		if autonomyRank(max) < 0 {
			result.Enabled, result.EffectiveAutonomy, result.Reason = false, AIAutonomyObserve, "invalid_tenant_autonomy_fails_closed"
			return result
		}
		if autonomyRank(capability.DefaultAutonomy) > autonomyRank(max) {
			result.Enabled, result.EffectiveAutonomy, result.Reason = true, max, "downgraded_by_tenant_max_autonomy"
			return result
		}
		result.Enabled, result.Reason = true, "allowed_by_tenant_policy"
		return result
	}
	return AIEffectiveCapability{Known: false, EffectiveAutonomy: AIAutonomyObserve, Reason: "unknown_purpose_or_action_denied"}
}

func BuildAICapabilityMatrix(policy TenantAIPolicy) AICapabilityMatrix {
	result := AICapabilityMatrix{SchemaVersion: AIAutonomyContractVersion, TenantID: policy.TenantID, MaxAutonomy: policy.EffectiveMaxAutonomy()}
	for _, capability := range AICapabilityCatalog() {
		result.Capabilities = append(result.Capabilities, ResolveAICapability(policy, capability.Purpose, capability.Action))
	}
	return result
}
