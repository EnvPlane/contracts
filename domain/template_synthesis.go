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

const TemplateSynthesisSchemaVersion = "1"

type TemplateSynthesisDecision struct {
	ResourceID      string  `json:"resourceId"`
	Action          string  `json:"action"`
	Strategy        string  `json:"strategy"`
	SourceType      string  `json:"sourceType"`
	SourceID        string  `json:"sourceId"`
	Path            string  `json:"path,omitempty"`
	Confidence      float64 `json:"confidence"`
	Reason          string  `json:"reason"`
	AutonomousApply bool    `json:"autonomousApply"`
}

type TemplateSynthesisInput struct {
	TenantID         string
	ProjectID        string
	ClusterID        string
	SourceScanID     string
	SourceNamespaces []string
	Now              time.Time
}

type EnvironmentTemplateSynthesis struct {
	SchemaVersion   string                      `json:"schemaVersion"`
	TenantID        string                      `json:"tenantId"`
	ProjectID       string                      `json:"projectId"`
	ClusterID       string                      `json:"clusterId"`
	SourceScanID    string                      `json:"sourceScanId"`
	Revision        EnvironmentTemplateRevision `json:"revision"`
	Graph           ServiceGraph                `json:"graph"`
	Decisions       []TemplateSynthesisDecision `json:"decisions"`
	Unresolved      []DependencyGraphIssue      `json:"unresolved,omitempty"`
	AutonomousApply bool                        `json:"autonomousApply"`
	Digest          string                      `json:"digest"`
}

func (s EnvironmentTemplateSynthesis) CanonicalDigest() (string, error) {
	c := s
	c.Digest = ""
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (s EnvironmentTemplateSynthesis) Validate() error {
	if s.SchemaVersion != TemplateSynthesisSchemaVersion || strings.TrimSpace(s.TenantID) == "" || strings.TrimSpace(s.ProjectID) == "" || strings.TrimSpace(s.SourceScanID) == "" || strings.TrimSpace(s.Digest) == "" {
		return fmt.Errorf("invalid template synthesis identity")
	}
	if s.Revision.TenantID != s.TenantID || s.Revision.ProjectID != s.ProjectID || s.Revision.SourceScanID != s.SourceScanID {
		return fmt.Errorf("template synthesis revision scope mismatch")
	}
	actual, err := s.CanonicalDigest()
	if err != nil {
		return err
	}
	if actual != s.Digest {
		return fmt.Errorf("template synthesis digest mismatch")
	}
	if err := s.Revision.Validate(); err != nil {
		return fmt.Errorf("template synthesis revision: %w", err)
	}
	for _, decision := range s.Decisions {
		if decision.ResourceID == "" || decision.SourceType == "" || decision.SourceID == "" || decision.Reason == "" {
			return fmt.Errorf("template synthesis decision %q is missing provenance", decision.ResourceID)
		}
		if decision.Confidence < 0 || decision.Confidence > 1 {
			return fmt.Errorf("template synthesis decision %q has invalid confidence", decision.ResourceID)
		}
	}
	return nil
}

func SortTemplateSynthesisDecisions(items []TemplateSynthesisDecision) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].ResourceID != items[j].ResourceID {
			return items[i].ResourceID < items[j].ResourceID
		}
		return items[i].Path < items[j].Path
	})
}
