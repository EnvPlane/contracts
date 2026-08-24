package domain

import "testing"

func TestClassifyAIRiskProhibitsCredentialActions(t *testing.T) {
	decision := ClassifyAIRisk(AIRiskInput{Action: "rotate", Credential: true})
	if !decision.Prohibited || decision.Approval != AIApprovalProhibited || decision.Autonomous {
		t.Fatalf("unexpected decision: %#v", decision)
	}
	if err := decision.Validate(); err != nil {
		t.Fatal(err)
	}
}
