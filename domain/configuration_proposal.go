package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const ConfigurationProposalSchemaVersion = "1"

type ConfigurationProposalRequest struct {
	Intent string `json:"intent"`
}

type ConfigurationProposalActionRequest struct {
	Action         string                `json:"action"`
	Intent         string                `json:"intent"`
	Proposal       ConfigurationProposal `json:"proposal"`
	JobID          string                `json:"jobId,omitempty"`
	IdempotencyKey string                `json:"idempotencyKey"`
}

type ConfigurationProposalActionResult struct {
	Status         string `json:"status"`
	Action         string `json:"action"`
	JobID          string `json:"jobId,omitempty"`
	IdempotencyKey string `json:"idempotencyKey"`
	ContextHash    string `json:"contextHash"`
	Message        string `json:"message,omitempty"`
}

type ConfigurationProposalFields struct {
	Mode           *EnvironmentMode             `json:"mode,omitempty"`
	ComponentRefs  []ConfigurationComponentRef  `json:"componentRefs,omitempty"`
	TTLHours       *int                         `json:"ttlHours,omitempty"`
	ResourceLimits *ConfigurationResourceLimits `json:"resourceLimits,omitempty"`
	HybridServices []string                     `json:"hybridServices,omitempty"`
}

type ConfigurationComponentRef struct {
	ComponentID string `json:"componentId"`
	Ref         string `json:"ref,omitempty"`
}

type ConfigurationResourceLimits struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type ConfigurationProposalRejectedField struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ConfigurationProposalDiff struct {
	Field  string `json:"field"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

type ConfigurationProposalRenderPreview struct {
	Mode            EnvironmentMode `json:"mode"`
	ComponentIDs    []string        `json:"componentIds,omitempty"`
	HybridServices  []string        `json:"hybridServices,omitempty"`
	TTLHours        int             `json:"ttlHours"`
	ResourceSummary []string        `json:"resourceSummary,omitempty"`
}

type ConfigurationProposal struct {
	SchemaVersion      string                               `json:"schemaVersion"`
	ProposalVersion    string                               `json:"proposalVersion"`
	TenantID           string                               `json:"tenantId"`
	ProjectID          string                               `json:"projectId"`
	Fields             ConfigurationProposalFields          `json:"fields"`
	Diff               []ConfigurationProposalDiff          `json:"diff,omitempty"`
	Warnings           []string                             `json:"warnings,omitempty"`
	RejectedFields     []ConfigurationProposalRejectedField `json:"rejectedFields,omitempty"`
	QuotaAllowed       bool                                 `json:"quotaAllowed"`
	QuotaReason        string                               `json:"quotaReason,omitempty"`
	Policy             PolicyDecision                       `json:"policy"`
	RenderPreview      ConfigurationProposalRenderPreview   `json:"renderPreview"`
	Valid              bool                                 `json:"valid"`
	ContextHash        string                               `json:"contextHash"`
	TargetStateVersion string                               `json:"targetStateVersion"`
}

func ConfigurationProposalContextHash(tenantID string, project Project, fields ConfigurationProposalFields) string {
	payload, _ := json.Marshal(struct {
		TenantID  string                      `json:"tenantId"`
		ProjectID string                      `json:"projectId"`
		UpdatedAt time.Time                   `json:"updatedAt"`
		Fields    ConfigurationProposalFields `json:"fields"`
	}{tenantID, project.ID, project.UpdatedAt.UTC(), (ConfigurationProposal{Fields: fields}).Deterministic().Fields})
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

func (f ConfigurationProposalFields) Validate() error {
	if f.Mode != nil && *f.Mode != ModeFull && *f.Mode != ModeHybrid {
		return fmt.Errorf("unsupported environment mode")
	}
	if f.TTLHours != nil && (*f.TTLHours < 1 || *f.TTLHours > 720) {
		return fmt.Errorf("ttlHours must be between 1 and 720")
	}
	seen := map[string]struct{}{}
	for _, ref := range f.ComponentRefs {
		if strings.TrimSpace(ref.ComponentID) == "" {
			return fmt.Errorf("componentId is required")
		}
		key := strings.TrimSpace(ref.ComponentID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate componentId %q", key)
		}
		seen[key] = struct{}{}
	}
	for _, service := range f.HybridServices {
		if strings.TrimSpace(service) == "" {
			return fmt.Errorf("hybrid service cannot be empty")
		}
	}
	return nil
}

func (p ConfigurationProposal) Deterministic() ConfigurationProposal {
	p.Fields.ComponentRefs = append([]ConfigurationComponentRef(nil), p.Fields.ComponentRefs...)
	sort.Slice(p.Fields.ComponentRefs, func(i, j int) bool {
		return p.Fields.ComponentRefs[i].ComponentID < p.Fields.ComponentRefs[j].ComponentID
	})
	p.Fields.HybridServices = append([]string(nil), p.Fields.HybridServices...)
	sort.Strings(p.Fields.HybridServices)
	sort.Slice(p.Diff, func(i, j int) bool { return p.Diff[i].Field < p.Diff[j].Field })
	sort.Strings(p.Warnings)
	sort.Slice(p.RejectedFields, func(i, j int) bool { return p.RejectedFields[i].Field < p.RejectedFields[j].Field })
	return p
}
