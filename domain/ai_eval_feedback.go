package domain

import (
	"errors"
	"strings"
	"time"
)

const AIEvalFeedbackSchemaVersion = "1"

type AIEvalFeedbackDimension string

const (
	AIEvalDimensionPlan     AIEvalFeedbackDimension = "plan"
	AIEvalDimensionAction   AIEvalFeedbackDimension = "action"
	AIEvalDimensionEvidence AIEvalFeedbackDimension = "evidence"
	AIEvalDimensionOutcome  AIEvalFeedbackDimension = "outcome"
)

type AIEvalConsent struct {
	OptedIn       bool      `json:"opted_in"`
	PolicyVersion string    `json:"policy_version"`
	ActorID       string    `json:"actor_id"`
	RecordedAt    time.Time `json:"recorded_at"`
}

type AIOutcomeFeedback struct {
	SchemaVersion string                  `json:"schema_version"`
	ID            string                  `json:"id"`
	RunID         string                  `json:"run_id"`
	TenantID      string                  `json:"tenant_id"`
	ProjectID     string                  `json:"project_id"`
	Dimension     AIEvalFeedbackDimension `json:"dimension"`
	Rating        string                  `json:"rating"`
	Comment       string                  `json:"comment,omitempty"`
	EvidenceIDs   []string                `json:"evidence_ids"`
	CreatedAt     time.Time               `json:"created_at"`
}

type AIEvalDatasetRecord struct {
	SchemaVersion string                `json:"schema_version"`
	ID            string                `json:"id"`
	TenantID      string                `json:"tenant_id"`
	RunID         string                `json:"run_id"`
	SnapshotHash  string                `json:"snapshot_hash"`
	PromptVersion string                `json:"prompt_version"`
	ModelVersion  string                `json:"model_version"`
	PolicyVersion string                `json:"policy_version"`
	Consent       AIEvalConsent         `json:"consent"`
	Provenance    []AIEvidenceReference `json:"provenance"`
	DeletedAt     *time.Time            `json:"deleted_at,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
}

type AISanitizedFact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type AIReplayRequest struct {
	SchemaVersion string              `json:"schema_version"`
	Dataset       AIEvalDatasetRecord `json:"dataset"`
	Facts         []AISanitizedFact   `json:"facts"`
	PromptVersion string              `json:"prompt_version"`
	ModelVersion  string              `json:"model_version"`
	PolicyVersion string              `json:"policy_version"`
}
type AIReplayResult struct {
	SchemaVersion     string `json:"schema_version"`
	SnapshotHash      string `json:"snapshot_hash"`
	Deterministic     bool   `json:"deterministic"`
	VersionCompatible bool   `json:"version_compatible"`
	Deleted           bool   `json:"deleted"`
}

func safeEvalText(value string) bool {
	return !strings.ContainsAny(value, "\r\n\x00") && len([]rune(value)) <= 1024 && !strings.Contains(strings.ToLower(value), "execute shell")
}

func (f AIOutcomeFeedback) Validate(tenantID string) error {
	if f.SchemaVersion != AIEvalFeedbackSchemaVersion || strings.TrimSpace(f.ID) == "" || strings.TrimSpace(f.RunID) == "" || f.TenantID != tenantID || strings.TrimSpace(f.ProjectID) == "" || len(f.EvidenceIDs) == 0 || f.CreatedAt.IsZero() || !safeEvalText(f.Comment) {
		return errors.New("AI outcome feedback is invalid or outside tenant scope")
	}
	switch f.Dimension {
	case AIEvalDimensionPlan, AIEvalDimensionAction, AIEvalDimensionEvidence, AIEvalDimensionOutcome:
	default:
		return errors.New("AI feedback dimension is invalid")
	}
	if f.Rating != "positive" && f.Rating != "negative" && f.Rating != "unknown" {
		return errors.New("AI feedback rating is invalid")
	}
	return nil
}
func (c AIEvalConsent) Validate() error {
	if !c.OptedIn || strings.TrimSpace(c.PolicyVersion) == "" || strings.TrimSpace(c.ActorID) == "" || c.RecordedAt.IsZero() {
		return errors.New("dataset consent is missing")
	}
	return nil
}
func (d AIEvalDatasetRecord) Validate(tenantID string, now time.Time) error {
	if d.SchemaVersion != AIEvalFeedbackSchemaVersion || strings.TrimSpace(d.ID) == "" || d.TenantID != tenantID || strings.TrimSpace(d.RunID) == "" || strings.TrimSpace(d.SnapshotHash) == "" || strings.TrimSpace(d.PromptVersion) == "" || strings.TrimSpace(d.ModelVersion) == "" || strings.TrimSpace(d.PolicyVersion) == "" || d.CreatedAt.IsZero() || len(d.Provenance) == 0 {
		return errors.New("eval dataset provenance is invalid")
	}
	if err := d.Consent.Validate(); err != nil {
		return err
	}
	for _, e := range d.Provenance {
		if e.TenantID != tenantID || e.Validate() != nil {
			return errors.New("eval provenance is outside tenant scope")
		}
	}
	if d.DeletedAt != nil && !d.DeletedAt.After(d.CreatedAt) {
		return errors.New("dataset deletion timestamp is invalid")
	}
	_ = now
	return nil
}
func (r AIReplayRequest) Validate(tenantID string, now time.Time) error {
	if r.SchemaVersion != AIEvalFeedbackSchemaVersion || len(r.Facts) > 256 || strings.TrimSpace(r.PromptVersion) == "" || strings.TrimSpace(r.ModelVersion) == "" || strings.TrimSpace(r.PolicyVersion) == "" {
		return errors.New("replay request is invalid")
	}
	if err := r.Dataset.Validate(tenantID, now); err != nil {
		return err
	}
	for _, f := range r.Facts {
		if strings.TrimSpace(f.Key) == "" || !safeEvalText(f.Key) || !safeEvalText(f.Value) {
			return errors.New("replay fact is unsafe or unbounded")
		}
	}
	return nil
}
