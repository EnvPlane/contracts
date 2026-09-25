package domain

import "testing"

func TestAIAutonomyDowngradesAndDisabledTenantFailsClosed(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	policy.Mode = AIPolicyExternal
	policy.MaxAutonomy = AIAutonomyRecommend
	capability := ResolveAICapability(policy, "environment_creation", "environment.create")
	if !capability.Known || !capability.Enabled || capability.EffectiveAutonomy != AIAutonomyRecommend || capability.Reason != "downgraded_by_tenant_max_autonomy" {
		t.Fatalf("unexpected downgrade: %#v", capability)
	}
	policy.Mode = AIPolicyDisabled
	capability = ResolveAICapability(policy, "environment_creation", "environment.create")
	if capability.Enabled || capability.Reason != "tenant_ai_disabled" {
		t.Fatalf("disabled tenant was not denied: %#v", capability)
	}
}

func TestAIAutonomyRejectsUnknownAndForbiddenActions(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	policy.Mode = AIPolicyExternal
	unknown := ResolveAICapability(policy, "unknown", "environment.delete")
	if unknown.Known || unknown.Enabled || unknown.Reason != "unknown_purpose_or_action_denied" {
		t.Fatalf("unknown capability was not denied: %#v", unknown)
	}
	forbidden := ResolveAICapability(policy, "approved_actions", "environment.delete")
	if !forbidden.Known || forbidden.Enabled || forbidden.ActionClass != AIActionForbidden {
		t.Fatalf("forbidden capability was not denied: %#v", forbidden)
	}
}

func TestAIAutonomyMatrixIsTenantScoped(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	policy.Mode = AIPolicyExternal
	matrix := BuildAICapabilityMatrix(policy)
	if matrix.TenantID != "tenant-a" || len(matrix.Capabilities) == 0 {
		t.Fatalf("invalid capability matrix: %#v", matrix)
	}
	if matrixForOther := BuildAICapabilityMatrix(DefaultTenantAIPolicy("tenant-b")); matrixForOther.TenantID != "tenant-b" {
		t.Fatalf("matrix leaked tenant scope: %#v", matrixForOther)
	}
}

func TestAIAutonomyMatrixHonorsRestrictedPolicyForImplementedPlanPurposes(t *testing.T) {
	policy := DefaultTenantAIPolicy("tenant-a")
	policy.Mode = AIPolicyExternal
	policy.MaxAutonomy = AIAutonomyApprovalRequired
	policy.Purposes = map[string]bool{"release.engineering": true, "incident.response": true, "security.compliance": true}
	for _, purpose := range []string{"release.engineering", "incident.response", "security.compliance"} {
		capability := ResolveAICapability(policy, purpose, map[string]string{
			"release.engineering": "release.plan",
			"incident.response":   "incident.plan",
			"security.compliance": "security.plan",
		}[purpose])
		if !capability.Known || !capability.Enabled || capability.EffectiveAutonomy != AIAutonomyApprovalRequired {
			t.Fatalf("implemented plan capability %q was not visible and enabled: %#v", purpose, capability)
		}
	}
	if capability := ResolveAICapability(policy, "scm.diagnosis", "scm.explain_or_propose"); capability.Enabled || capability.Reason != "purpose_disabled_by_tenant_policy" {
		t.Fatalf("restricted policy did not disable SCM diagnosis: %#v", capability)
	}
}
