package domain

import (
	"fmt"
	"strings"
	"time"
)

const UsageEventSchemaVersion = "1"

const (
	UsageMetricManagedCluster     = "managed_cluster"
	UsageMetricEnvironmentActive  = "environment_active"
	UsageMetricOperatorMembership = "operator_membership"
	UsageMetricAIInputTokens      = "ai_input_tokens"
	UsageMetricAIOutputTokens     = "ai_output_tokens"
	UsageMetricAICostMicros       = "ai_cost_micros"
)

type UsageEvent struct {
	ID             string    `json:"id"`
	IdempotencyKey string    `json:"idempotencyKey"`
	TenantID       string    `json:"tenantId"`
	Metric         string    `json:"metric"`
	Quantity       int64     `json:"quantity"`
	OccurredAt     time.Time `json:"occurredAt"`
	SourceResource string    `json:"sourceResource"`
	SchemaVersion  string    `json:"schemaVersion"`
}

func (e UsageEvent) Validate() error {
	if strings.TrimSpace(e.TenantID) == "" || strings.TrimSpace(e.IdempotencyKey) == "" || strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("usage event identity is required")
	}
	if e.Quantity == 0 {
		return fmt.Errorf("usage event quantity must not be zero")
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("usage event occurred_at is required")
	}
	if e.SchemaVersion != UsageEventSchemaVersion {
		return fmt.Errorf("unsupported usage event schema version %q", e.SchemaVersion)
	}
	switch e.Metric {
	case UsageMetricManagedCluster, UsageMetricEnvironmentActive, UsageMetricOperatorMembership, UsageMetricAIInputTokens, UsageMetricAIOutputTokens, UsageMetricAICostMicros:
	default:
		return fmt.Errorf("unsupported usage metric %q", e.Metric)
	}
	resource := strings.TrimSpace(e.SourceResource)
	if resource == "" || strings.Contains(resource, "://") || strings.ContainsAny(resource, "\r\n@") {
		return fmt.Errorf("usage source resource must be an opaque identifier")
	}
	return nil
}
