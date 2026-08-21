package domain

import (
	"fmt"
	"sort"
	"strings"
)

const AIPolicySchemaVersion = "1"

type AIPolicyMode string

const (
	AIPolicyDisabled   AIPolicyMode = "disabled"
	AIPolicyExternal   AIPolicyMode = "external"
	AIPolicySelfHosted AIPolicyMode = "self_hosted"
)

type AIRetentionMode string

const (
	AIRetentionMetadataOnly AIRetentionMode = "metadata_only"
	AIRetentionNone         AIRetentionMode = "none"
)

type TenantAIPolicy struct {
	SchemaVersion            string          `json:"schemaVersion"`
	TenantID                 string          `json:"tenantId"`
	Mode                     AIPolicyMode    `json:"mode"`
	AllowedRegions           []string        `json:"allowedRegions,omitempty"`
	AllowedEndpoints         []string        `json:"allowedEndpoints,omitempty"`
	AllowedModels            []string        `json:"allowedModels,omitempty"`
	RetentionMode            AIRetentionMode `json:"retentionMode"`
	MaxContextClassification string          `json:"maxContextClassification"`
	Purposes                 map[string]bool `json:"purposes,omitempty"`
}

func DefaultTenantAIPolicy(tenantID string) TenantAIPolicy {
	return TenantAIPolicy{SchemaVersion: AIPolicySchemaVersion, TenantID: strings.TrimSpace(tenantID), Mode: AIPolicyDisabled, RetentionMode: AIRetentionMetadataOnly, MaxContextClassification: "safe_metadata", Purposes: map[string]bool{}}
}

func (p TenantAIPolicy) Validate() error {
	if p.SchemaVersion != AIPolicySchemaVersion || strings.TrimSpace(p.TenantID) == "" {
		return fmt.Errorf("invalid AI policy identity or schema")
	}
	if p.Mode != AIPolicyDisabled && p.Mode != AIPolicyExternal && p.Mode != AIPolicySelfHosted {
		return fmt.Errorf("unsupported AI policy mode")
	}
	if p.RetentionMode != AIRetentionMetadataOnly && p.RetentionMode != AIRetentionNone {
		return fmt.Errorf("unsupported AI retention mode")
	}
	if strings.TrimSpace(p.MaxContextClassification) == "" || p.MaxContextClassification != "safe_metadata" {
		return fmt.Errorf("AI context classification cannot exceed safe_metadata")
	}
	for purpose := range p.Purposes {
		if purpose != "diagnosis" && purpose != "bootstrap_troubleshooting" && purpose != "bootstrap.scan" && purpose != "configuration_proposal" && purpose != "finops_explanation" {
			return fmt.Errorf("unsupported AI purpose %q", purpose)
		}
	}
	return nil
}

func (p TenantAIPolicy) Deterministic() TenantAIPolicy {
	p.AllowedRegions = append([]string(nil), p.AllowedRegions...)
	p.AllowedEndpoints = append([]string(nil), p.AllowedEndpoints...)
	p.AllowedModels = append([]string(nil), p.AllowedModels...)
	sort.Strings(p.AllowedRegions)
	sort.Strings(p.AllowedEndpoints)
	sort.Strings(p.AllowedModels)
	return p
}

func (p TenantAIPolicy) PurposeEnabled(purpose string) bool {
	if p.Mode == AIPolicyDisabled {
		return false
	}
	if len(p.Purposes) == 0 {
		return true
	}
	return p.Purposes[purpose]
}
