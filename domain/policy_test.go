package domain

import "testing"

func TestBuiltinPolicyBundleIsVersionedAndDeterministic(t *testing.T) {
	bundle := BuiltinPolicyBundle("2026.08")
	if err := bundle.Validate(); err != nil {
		t.Fatal(err)
	}
	first := bundle.Deterministic()
	second := bundle.Deterministic()
	if len(first.Rules) != len(second.Rules) || first.Rules[0].ID != second.Rules[0].ID {
		t.Fatalf("bundle is not deterministic: %#v %#v", first, second)
	}
	if first.SchemaVersion != PolicyBundleSchemaVersion || first.BundleVersion != "2026.08" {
		t.Fatalf("unexpected versions: %#v", first)
	}
}

func TestPolicyBundleRejectsDuplicateAndUnknownRules(t *testing.T) {
	bundle := PolicyBundle{SchemaVersion: PolicyBundleSchemaVersion, BundleVersion: "1", Rules: []PolicyRule{
		{ID: "same", Action: "create_environment", Check: "quota_allowed", Effect: PolicyDeny},
		{ID: "same", Action: "create_environment", Check: "quota_allowed", Effect: PolicyDeny},
	}}
	if err := bundle.Validate(); err == nil {
		t.Fatal("duplicate rule accepted")
	}
	bundle.Rules[1].ID = "unknown"
	bundle.Rules[1].Check = "arbitrary_runtime"
	if err := bundle.Validate(); err == nil {
		t.Fatal("unknown check accepted")
	}
}
