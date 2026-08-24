package domain

import (
	"errors"
	"strings"
	"time"
)

const AIRetrievalSchemaVersion = "1"

const (
	AIRetrievalEnvironment = "environment"
	AIRetrievalEvent       = "kubernetes_event"
	AIRetrievalJob         = "job"
	AIRetrievalScan        = "resource_scan"
	AIRetrievalGitOps      = "gitops_status"
	AIRetrievalRunbook     = "approved_runbook"
)

type AIRetrievalQuery struct {
	SchemaVersion string   `json:"schemaVersion"`
	TenantID      string   `json:"tenantId"`
	ProjectID     string   `json:"projectId"`
	SubjectType   string   `json:"subjectType"`
	SubjectID     string   `json:"subjectId"`
	Sources       []string `json:"sources"`
	MaxResults    int      `json:"maxResults"`
	RequireFresh  bool     `json:"requireFresh"`
	FreshFor      string   `json:"freshFor"`
}

type AIRetrievedEvidence struct {
	SchemaVersion string              `json:"schemaVersion"`
	EvidenceID    string              `json:"evidenceId"`
	Reference     AIEvidenceReference `json:"reference"`
	Fields        map[string]string   `json:"fields"`
	ObservedAt    time.Time           `json:"observedAt"`
	ExpiresAt     time.Time           `json:"expiresAt"`
	Stale         bool                `json:"stale"`
}

type AIShortTermMemoryFact struct {
	SchemaVersion string    `json:"schemaVersion"`
	ID            string    `json:"id"`
	RunID         string    `json:"runId"`
	TenantID      string    `json:"tenantId"`
	ProjectID     string    `json:"projectId"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	EvidenceID    string    `json:"evidenceId"`
	ObservedAt    time.Time `json:"observedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

type AIMaterialConclusion struct {
	SchemaVersion string   `json:"schemaVersion"`
	Summary       string   `json:"summary"`
	EvidenceIDs   []string `json:"evidenceIds"`
}

func (q AIRetrievalQuery) Validate() error {
	if q.SchemaVersion != AIRetrievalSchemaVersion || strings.TrimSpace(q.TenantID) == "" || strings.TrimSpace(q.ProjectID) == "" || strings.TrimSpace(q.SubjectType) == "" || strings.TrimSpace(q.SubjectID) == "" || len(q.Sources) == 0 || len(q.Sources) > 8 || q.MaxResults <= 0 || q.MaxResults > 256 {
		return errors.New("AI retrieval query is invalid or unbounded")
	}
	for _, source := range q.Sources {
		switch source {
		case AIRetrievalEnvironment, AIRetrievalEvent, AIRetrievalJob, AIRetrievalScan, AIRetrievalGitOps, AIRetrievalRunbook:
		default:
			return errors.New("AI retrieval source is not allowlisted")
		}
	}
	return nil
}

func (e AIRetrievedEvidence) Validate(now time.Time) error {
	if e.SchemaVersion != AIRetrievalSchemaVersion || strings.TrimSpace(e.EvidenceID) == "" || e.Reference.TenantID == "" || e.Reference.SourceType == "" || e.Reference.SourceID == "" || len(e.Fields) > 32 || e.ObservedAt.IsZero() || e.ExpiresAt.IsZero() || !e.ExpiresAt.After(e.ObservedAt) {
		return errors.New("AI retrieved evidence is invalid")
	}
	if e.Reference.TenantID != strings.TrimSpace(e.Reference.TenantID) {
		return errors.New("AI evidence tenant is invalid")
	}
	if !now.IsZero() && e.ExpiresAt.Before(now.UTC()) && !e.Stale {
		return errors.New("expired evidence must be marked stale")
	}
	for key, value := range e.Fields {
		if strings.TrimSpace(key) == "" || len(key) > 128 || len(value) > 1024 {
			return errors.New("AI evidence field is invalid or unbounded")
		}
	}
	return nil
}

func (f AIShortTermMemoryFact) Validate(now time.Time) error {
	if f.SchemaVersion != AIRetrievalSchemaVersion || strings.TrimSpace(f.ID) == "" || strings.TrimSpace(f.RunID) == "" || strings.TrimSpace(f.TenantID) == "" || strings.TrimSpace(f.ProjectID) == "" || strings.TrimSpace(f.Key) == "" || strings.TrimSpace(f.Value) == "" || strings.TrimSpace(f.EvidenceID) == "" || f.ObservedAt.IsZero() || f.ExpiresAt.IsZero() || !f.ExpiresAt.After(f.ObservedAt) {
		return errors.New("AI memory fact is invalid")
	}
	if len(f.Key) > 128 || len(f.Value) > 1024 || (!now.IsZero() && f.ExpiresAt.Before(now.UTC())) {
		return errors.New("AI memory fact is expired or unbounded")
	}
	return nil
}

func (c AIMaterialConclusion) Validate(allowed map[string]struct{}) error {
	if c.SchemaVersion != AIRetrievalSchemaVersion || strings.TrimSpace(c.Summary) == "" || len(c.EvidenceIDs) == 0 {
		return errors.New("AI material conclusion requires evidence")
	}
	seen := map[string]struct{}{}
	for _, id := range c.EvidenceIDs {
		if strings.TrimSpace(id) == "" {
			return errors.New("AI conclusion evidence ID is empty")
		}
		if _, duplicate := seen[id]; duplicate {
			return errors.New("AI conclusion evidence IDs are duplicated")
		}
		if _, ok := allowed[id]; !ok {
			return errors.New("AI conclusion references unavailable evidence")
		}
		seen[id] = struct{}{}
	}
	return nil
}
