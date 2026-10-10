package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSecretMaterializationOutputUIDCompatibility(t *testing.T) {
	p, err := CompileSecretMaterializationPlan("tenant", "project", "environment", "revision", "sha256:template", "target", []SecretStrategyConfig{{ID: "generated", Strategy: SecretStrategyGenerated, Generator: "random-password-v1", TargetNamespace: "target", TargetName: "generated"}}, "sha256:input", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	key, err := SecretMaterializationIdempotencyKey(p.TenantID, p.ProjectID, p.EnvironmentID, p.TemplateDigest, p.TargetNamespace, "generated", SecretOperationMaterialize)
	if err != nil {
		t.Fatal(err)
	}
	p.Results = []SecretMaterializationItemResult{{ItemID: "generated", Strategy: SecretStrategyGenerated, TargetNamespace: "target", TargetName: "generated", Operation: SecretOperationMaterialize, IdempotencyKey: key, InputDigest: p.InputDigest, Status: SecretItemReady, Attempt: 1, StartedAt: time.Unix(1, 0), FinishedAt: time.Unix(2, 0)}}
	if err := p.Validate(); err != nil {
		t.Fatal("legacy result without UID rejected", err)
	}
	before, _ := p.CanonicalDigest()
	p.Results[0].OutputUID = "generated-uid"
	after, _ := p.CanonicalDigest()
	if before != after {
		t.Fatal("result UID changed immutable Secret plan digest")
	}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(AgentSecretMaterializationResult{Items: p.Results})
	if !strings.Contains(string(encoded), `"outputUid":"generated-uid"`) {
		t.Fatal("Agent result dropped UID")
	}
	var decoded AgentSecretMaterializationResult
	if err := json.Unmarshal(encoded, &decoded); err != nil || decoded.Items[0].OutputUID != "generated-uid" {
		t.Fatal("UID round-trip failed", err)
	}
	p.Results[0].OutputUID = "uid;id"
	if err := p.Validate(); err == nil {
		t.Fatal("malicious UID accepted")
	}
}
