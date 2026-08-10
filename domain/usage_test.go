package domain

import (
	"testing"
	"time"
)

func TestUsageEventRejectsSensitiveSourceShape(t *testing.T) {
	event := UsageEvent{ID: "evt", IdempotencyKey: "key", TenantID: "tenant", Metric: UsageMetricEnvironmentActive, Quantity: 1, OccurredAt: time.Now(), SourceResource: "env-1", SchemaVersion: UsageEventSchemaVersion}
	if err := event.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"https://repo.example", "branch@example", "env\nsecret"} {
		event.SourceResource = source
		if err := event.Validate(); err == nil {
			t.Fatalf("source %q accepted", source)
		}
	}
}
