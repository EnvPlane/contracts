package domain

import (
	"fmt"
	"sort"
	"strings"
)

const PlanSchemaVersion = "1"

const (
	FeatureAuthOIDC          = "auth.oidc"
	FeatureAuthSAML          = "auth.saml"
	FeatureIdentitySCIM      = "identity.scim"
	FeatureRBACGranular      = "rbac.granular"
	FeatureGitOpsFlux        = "gitops.flux"
	FeatureGitOpsArgo        = "gitops.argo"
	FeatureFinOpsAllocation  = "finops.allocation"
	FeaturePolicyCustom      = "policy.custom"
	FeatureFleetUpgradeWaves = "fleet.upgrade_waves"
	FeatureSupportSLA        = "support.sla"
	FeatureAuditSIEM        = "audit.siem_export"
	FeatureAIDiagnosis     = "ai.diagnosis"
	FeatureAIBootstrap     = "ai.bootstrap"
	FeatureAIConfiguration = "ai.configuration"
	FeatureAIEnvironment   = "ai.environment_create"
	FeatureAIFinOps        = "ai.finops"
	FeatureAIApproved      = "ai.approved_actions"
)

const (
	LimitProjectsMax           = "projects.max"
	LimitManagedClustersMax    = "clusters.managed.max"
	LimitActiveEnvironmentsMax = "environments.active.max"
	LimitEnvironmentTTLHours   = "environment.ttl.max_hours"
	LimitEnvironmentPinHours   = "environment.pin.max_hours"
	LimitOperatorsMax          = "users.operators.max"
	LimitAuditRetentionDays    = "audit.retention_days"
	LimitAIRunsConcurrent      = "ai.runs.max_concurrent"
	LimitAIRunsRequests        = "ai.runs.max_requests"
	LimitAIContextBytes        = "ai.context.max_bytes"
	LimitAIOutputTokens        = "ai.output.max_tokens"
)

var knownPlanFeatureKeys = map[string]struct{}{
	FeatureAuthOIDC: {}, FeatureAuthSAML: {}, FeatureIdentitySCIM: {}, FeatureRBACGranular: {},
	FeatureGitOpsFlux: {}, FeatureGitOpsArgo: {}, FeatureFinOpsAllocation: {}, FeaturePolicyCustom: {},
	FeatureFleetUpgradeWaves: {}, FeatureSupportSLA: {},
	FeatureAuditSIEM: {},
	FeatureAIDiagnosis: {}, FeatureAIBootstrap: {}, FeatureAIConfiguration: {}, FeatureAIEnvironment: {}, FeatureAIFinOps: {}, FeatureAIApproved: {},
	// These aliases are retained for existing entitlement and quota callers.
	"projects": {}, "environments": {}, "gitops": {}, "helmDirect": {}, "audit": {},
}

var knownPlanLimitKeys = map[string]struct{}{
	LimitProjectsMax: {}, LimitManagedClustersMax: {}, LimitActiveEnvironmentsMax: {},
	LimitEnvironmentTTLHours: {}, LimitEnvironmentPinHours: {}, LimitOperatorsMax: {}, LimitAuditRetentionDays: {},
	LimitAIRunsConcurrent: {}, LimitAIRunsRequests: {}, LimitAIContextBytes: {}, LimitAIOutputTokens: {},
	// These aliases are retained for existing entitlement and quota callers.
	"maxProjects": {}, "maxRemoteClusters": {}, "maxActiveEnvironments": {}, "maxMembers": {},
	"maxTTLHours": {}, "maxPinDays": {},
}

type PlanDefinition struct {
	ID               string           `json:"id"`
	SchemaVersion    string           `json:"schemaVersion"`
	EffectiveVersion string           `json:"effectiveVersion"`
	Features         map[string]bool  `json:"features"`
	Limits           map[string]int64 `json:"limits"`
}

type PlanCatalog struct {
	SchemaVersion string           `json:"schemaVersion"`
	Plans         []PlanDefinition `json:"plans"`
}

func CommunityFreePlanCatalog() PlanCatalog {
	return PlanCatalog{
		SchemaVersion: PlanSchemaVersion,
		Plans: []PlanDefinition{
			communityPlan("1.0.0"),
			freePlan("1.0.0"),
		},
	}
}

func communityPlan(version string) PlanDefinition {
	return PlanDefinition{ID: "community", SchemaVersion: PlanSchemaVersion, EffectiveVersion: version,
		Features: map[string]bool{
			FeatureAuthOIDC: true, FeatureAuthSAML: false, FeatureIdentitySCIM: false, FeatureRBACGranular: false,
			FeatureGitOpsFlux: true, FeatureGitOpsArgo: false, FeatureFinOpsAllocation: false, FeaturePolicyCustom: false,
			FeatureFleetUpgradeWaves: false, FeatureSupportSLA: false,
			FeatureAuditSIEM: true,
			"projects": true, "environments": true, "gitops": true, "helmDirect": true, "audit": true,
		}, Limits: map[string]int64{
			LimitProjectsMax: 10, LimitManagedClustersMax: 3, LimitActiveEnvironmentsMax: 25,
			LimitEnvironmentTTLHours: 720, LimitEnvironmentPinHours: 720, LimitOperatorsMax: 10, LimitAuditRetentionDays: 30,
			"maxProjects": 10, "maxRemoteClusters": 3, "maxActiveEnvironments": 25, "maxMembers": 10,
			"maxTTLHours": 720, "maxPinDays": 30,
		}}
}

func freePlan(version string) PlanDefinition {
	return PlanDefinition{ID: "free", SchemaVersion: PlanSchemaVersion, EffectiveVersion: version,
		Features: map[string]bool{
			FeatureAuthOIDC: true, FeatureAuthSAML: false, FeatureIdentitySCIM: false, FeatureRBACGranular: false,
			FeatureGitOpsFlux: true, FeatureGitOpsArgo: false, FeatureFinOpsAllocation: false, FeaturePolicyCustom: false,
			FeatureFleetUpgradeWaves: false, FeatureSupportSLA: false,
			FeatureAuditSIEM: false,
			"projects": true, "environments": true, "gitops": true, "helmDirect": true, "audit": true,
		}, Limits: map[string]int64{
			LimitProjectsMax: 3, LimitManagedClustersMax: 1, LimitActiveEnvironmentsMax: 2,
			LimitEnvironmentTTLHours: 72, LimitEnvironmentPinHours: 168, LimitOperatorsMax: 3, LimitAuditRetentionDays: 7,
			"maxProjects": 3, "maxRemoteClusters": 1, "maxActiveEnvironments": 2, "maxMembers": 3,
			"maxTTLHours": 72, "maxPinDays": 7,
		}}
}

func (c PlanCatalog) Validate() error {
	if c.SchemaVersion != PlanSchemaVersion {
		return fmt.Errorf("unsupported plan catalog schema version %q", c.SchemaVersion)
	}
	seen := map[string]struct{}{}
	for _, plan := range c.Plans {
		if strings.TrimSpace(plan.ID) == "" || strings.TrimSpace(plan.EffectiveVersion) == "" {
			return fmt.Errorf("plan id and effective version are required")
		}
		if plan.SchemaVersion != c.SchemaVersion {
			return fmt.Errorf("plan %q schema version mismatch", plan.ID)
		}
		if _, ok := seen[plan.ID]; ok {
			return fmt.Errorf("duplicate plan %q", plan.ID)
		}
		seen[plan.ID] = struct{}{}
		for key, limit := range plan.Limits {
			if strings.TrimSpace(key) == "" || limit < 0 {
				return fmt.Errorf("invalid limit %q in plan %q", key, plan.ID)
			}
			if _, ok := knownPlanLimitKeys[key]; !ok {
				return fmt.Errorf("unknown limit key %q in plan %q", key, plan.ID)
			}
		}
		for key := range plan.Features {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("empty feature key in plan %q", plan.ID)
			}
			if _, ok := knownPlanFeatureKeys[key]; !ok {
				return fmt.Errorf("unknown feature key %q in plan %q", key, plan.ID)
			}
		}
	}
	return nil
}

func (c PlanCatalog) Deterministic() PlanCatalog {
	copyCatalog := PlanCatalog{SchemaVersion: c.SchemaVersion, Plans: make([]PlanDefinition, len(c.Plans))}
	for i, plan := range c.Plans {
		copyCatalog.Plans[i] = PlanDefinition{ID: plan.ID, SchemaVersion: plan.SchemaVersion, EffectiveVersion: plan.EffectiveVersion,
			Features: make(map[string]bool, len(plan.Features)), Limits: make(map[string]int64, len(plan.Limits))}
		for key, value := range plan.Features {
			copyCatalog.Plans[i].Features[key] = value
		}
		for key, value := range plan.Limits {
			copyCatalog.Plans[i].Limits[key] = value
		}
	}
	sort.Slice(copyCatalog.Plans, func(i, j int) bool { return copyCatalog.Plans[i].ID < copyCatalog.Plans[j].ID })
	return copyCatalog
}
