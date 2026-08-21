package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const EnvironmentCreateProposalSchemaVersion = "1"

type EnvironmentCreateProposalRequest struct {
	Intent        string   `json:"intent"`
	Branch        string   `json:"branch,omitempty"`
	PullRequestID string   `json:"pullRequestId,omitempty"`
	ChangedPaths  []string `json:"changedPaths,omitempty"`
}

type EnvironmentCreateProposalFields struct {
	Mode           *EnvironmentMode             `json:"mode,omitempty"`
	Components     []ComponentOverride          `json:"components,omitempty"`
	Branch         string                       `json:"branch,omitempty"`
	PullRequestID  string                       `json:"pullRequestId,omitempty"`
	TTLHours       *int                         `json:"ttlHours,omitempty"`
	Pinned         *bool                        `json:"pinned,omitempty"`
	ResourceLimits *ConfigurationResourceLimits `json:"resourceLimits,omitempty"`
}

type EnvironmentCreateProposal struct {
	SchemaVersion      string                               `json:"schemaVersion"`
	ProposalVersion    string                               `json:"proposalVersion"`
	Kind               string                               `json:"kind"`
	TenantID           string                               `json:"tenantId"`
	ProjectID          string                               `json:"projectId"`
	Fields             EnvironmentCreateProposalFields      `json:"fields"`
	RejectedFields     []ConfigurationProposalRejectedField `json:"rejectedFields,omitempty"`
	Warnings           []string                             `json:"warnings,omitempty"`
	QuotaAllowed       bool                                 `json:"quotaAllowed"`
	Policy             PolicyDecision                       `json:"policy"`
	ApprovalRequired   bool                                 `json:"approvalRequired"`
	Valid              bool                                 `json:"valid"`
	ContextHash        string                               `json:"contextHash"`
	TargetStateVersion string                               `json:"targetStateVersion"`
	QuotaGuidance      *EnvironmentCreateQuotaGuidance     `json:"quotaGuidance,omitempty"`
}

type EnvironmentCreateQuotaImpact struct {
	Resource      string `json:"resource"`
	Limit         int64  `json:"limit,omitempty"`
	Requested     int64  `json:"requested,omitempty"`
	Allowed       bool   `json:"allowed"`
	Known         bool   `json:"known"`
	Explanation   string `json:"explanation"`
}

type EnvironmentCreateCostGuidance struct {
	Currency          string `json:"currency"`
	CurrentMinorUnits int64  `json:"currentMinorUnits,omitempty"`
	CurrentKnown      bool   `json:"currentKnown"`
	ProjectedMinorUnits int64 `json:"projectedMinorUnits,omitempty"`
	ProjectedKnown     bool  `json:"projectedKnown"`
	Explanation        string `json:"explanation"`
}

type EnvironmentCreateQuotaGuidance struct {
	SchemaVersion string                         `json:"schemaVersion"`
	Quota         []EnvironmentCreateQuotaImpact `json:"quota"`
	Cost          EnvironmentCreateCostGuidance  `json:"cost"`
	DataSufficient bool                          `json:"dataSufficient"`
}

type EnvironmentCreateProposalActionRequest struct {
	Proposal       EnvironmentCreateProposal `json:"proposal"`
	ApprovalID     string                    `json:"approvalId"`
	IdempotencyKey string                    `json:"idempotencyKey"`
}

type EnvironmentCreateProposalActionResult struct {
	Status         string `json:"status"`
	ApprovalID     string `json:"approvalId,omitempty"`
	EnvironmentID  string `json:"environmentId,omitempty"`
	ContextHash    string `json:"contextHash"`
	IdempotencyKey string `json:"idempotencyKey"`
	Message        string `json:"message,omitempty"`
	Failure        *EnvironmentCreateFailure `json:"failure,omitempty"`
}

type EnvironmentCreateFailureCategory string

const (
	EnvironmentCreateFailureRender         EnvironmentCreateFailureCategory = "render_error"
	EnvironmentCreateFailureQuota          EnvironmentCreateFailureCategory = "quota_block"
	EnvironmentCreateFailureInvalidOverride EnvironmentCreateFailureCategory = "invalid_override"
	EnvironmentCreateFailureTimeout        EnvironmentCreateFailureCategory = "job_timeout"
	EnvironmentCreateFailureUnknown        EnvironmentCreateFailureCategory = "unknown"
)

type EnvironmentCreateProposalDiff struct {
	Field  string `json:"field"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
	Reason string `json:"reason"`
}

type EnvironmentCreateFailure struct {
	Category        EnvironmentCreateFailureCategory `json:"category"`
	Diagnosis       AIDiagnosisResult                 `json:"diagnosis"`
	FollowUp        *EnvironmentCreateProposal       `json:"followUp,omitempty"`
	Diff            []EnvironmentCreateProposalDiff  `json:"diff,omitempty"`
	SuggestionsLeft int                               `json:"suggestionsLeft"`
}

func (f EnvironmentCreateProposalFields) Deterministic() EnvironmentCreateProposalFields {
	f.Components = append([]ComponentOverride(nil), f.Components...)
	sort.Slice(f.Components, func(i, j int) bool { return f.Components[i].ComponentID < f.Components[j].ComponentID })
	return f
}

func (f EnvironmentCreateProposalFields) Validate() error {
	if f.Mode != nil && *f.Mode != ModeFull && *f.Mode != ModeHybrid {
		return errors.New("unsupported environment mode")
	}
	if strings.TrimSpace(f.Branch) != "" && strings.TrimSpace(f.PullRequestID) != "" {
		return errors.New("branch and pullRequestId cannot both be set")
	}
	if strings.TrimSpace(f.Branch) == "" && strings.TrimSpace(f.PullRequestID) == "" {
		return errors.New("branch or pullRequestId is required")
	}
	if f.TTLHours != nil && (*f.TTLHours < 1 || *f.TTLHours > 720) {
		return errors.New("ttlHours must be between 1 and 720")
	}
	seen := map[string]struct{}{}
	for _, component := range f.Components {
		if strings.TrimSpace(component.ComponentID) == "" {
			return errors.New("componentId is required")
		}
		if _, ok := seen[component.ComponentID]; ok {
			return fmt.Errorf("duplicate componentId %q", component.ComponentID)
		}
		seen[component.ComponentID] = struct{}{}
		if strings.TrimSpace(component.Branch) != "" && strings.TrimSpace(component.PullRequestID) != "" {
			return errors.New("component branch and pullRequestId cannot both be set")
		}
	}
	if f.ResourceLimits != nil {
		if strings.TrimSpace(f.ResourceLimits.CPU) == "" && strings.TrimSpace(f.ResourceLimits.Memory) == "" {
			return errors.New("resource limits cannot be empty")
		}
	}
	return nil
}

func EnvironmentCreateProposalContextHash(tenantID string, project Project, fields EnvironmentCreateProposalFields) string {
	payload, _ := json.Marshal(struct {
		TenantID, ProjectID, TargetState string
		Fields                           EnvironmentCreateProposalFields
	}{tenantID, project.ID, project.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"), fields.Deterministic()})
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}
