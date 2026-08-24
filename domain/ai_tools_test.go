package domain

import "testing"

func TestAIToolDescriptorRejectsForbiddenAndReadOnlyWriteTools(t *testing.T) {
	base := AIToolDescriptor{SchemaVersion: AIToolRegistrySchemaVersion, Name: "refresh_environment_status", Risk: AIRiskLow, ActionClass: AIActionReadOnly, ReadOnly: true}
	if err := base.Validate(); err != nil { t.Fatal(err) }
	base.ActionClass = AIActionForbidden
	if err := base.Validate(); err == nil { t.Fatal("forbidden tool descriptor accepted") }
	base.ActionClass = AIActionApprovedWrite
	if err := base.Validate(); err == nil { t.Fatal("read-only write tool descriptor accepted") }
}
