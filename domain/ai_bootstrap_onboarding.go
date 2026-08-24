package domain

import (
	"errors"
	"strings"
	"time"
)

const AIBootstrapOnboardingSchemaVersion = "1"

type AIBootstrapDiscovery struct {
	RepositoryProvider string   `json:"repositoryProvider,omitempty"`
	Repository         string   `json:"repository,omitempty"`
	DefaultBranch      string   `json:"defaultBranch,omitempty"`
	Branches           []string `json:"branches,omitempty"`
	ClusterID          string   `json:"clusterId,omitempty"`
	Namespaces         []string `json:"namespaces,omitempty"`
	KubernetesVersion  string   `json:"kubernetesVersion,omitempty"`
	GitOpsBackend      string   `json:"gitOpsBackend,omitempty"`
	GitOpsRepository   string   `json:"gitOpsRepository,omitempty"`
	SecretStrategy     string   `json:"secretStrategy,omitempty"`
}

type AIBootstrapFieldSuggestion struct {
	Field          string                `json:"field"`
	Value          string                `json:"value"`
	Evidence       []AIEvidenceReference `json:"evidence"`
	Confidence     AIConfidence          `json:"confidence"`
	RequiresReview bool                  `json:"requiresReview"`
	Untrusted      bool                  `json:"untrusted"`
}

type AIBootstrapOnboardingPlan struct {
	SchemaVersion       string                       `json:"schemaVersion"`
	PlanID              string                       `json:"planId"`
	TenantID            string                       `json:"tenantId"`
	ProjectID           string                       `json:"projectId"`
	BootstrapSessionID  string                       `json:"bootstrapSessionId"`
	ContextHash         string                       `json:"contextHash"`
	Discovery           AIBootstrapDiscovery         `json:"discovery"`
	Suggestions         []AIBootstrapFieldSuggestion `json:"suggestions"`
	BlockedReasons      []string                     `json:"blockedReasons,omitempty"`
	CredentialsRequired bool                         `json:"credentialsRequired"`
	ApprovalRequired    bool                         `json:"approvalRequired"`
	ManualAuthoritative bool                         `json:"manualAuthoritative"`
	ReadyForReview      bool                         `json:"readyForReview"`
	GeneratedAt         time.Time                    `json:"generatedAt"`
}

func (d AIBootstrapDiscovery) Validate() error {
	for _, value := range []string{d.RepositoryProvider, d.Repository, d.DefaultBranch, d.ClusterID, d.KubernetesVersion, d.GitOpsBackend, d.GitOpsRepository, d.SecretStrategy} {
		if len(value) > 512 || strings.ContainsAny(value, "\r\n") {
			return errors.New("bootstrap discovery contains an invalid value")
		}
	}
	if len(d.Branches) > 128 || len(d.Namespaces) > 128 {
		return errors.New("bootstrap discovery is unbounded")
	}
	return nil
}

func (p AIBootstrapOnboardingPlan) Validate() error {
	if p.SchemaVersion != AIBootstrapOnboardingSchemaVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.BootstrapSessionID) == "" || strings.TrimSpace(p.ContextHash) == "" || p.GeneratedAt.IsZero() || !p.ManualAuthoritative {
		return errors.New("bootstrap onboarding plan identity or authority is invalid")
	}
	if err := p.Discovery.Validate(); err != nil {
		return err
	}
	if len(p.Suggestions) > 64 || len(p.BlockedReasons) > 32 {
		return errors.New("bootstrap onboarding plan is unbounded")
	}
	for _, suggestion := range p.Suggestions {
		if strings.TrimSpace(suggestion.Field) == "" || len(suggestion.Field) > 128 || len(suggestion.Value) > 1024 || strings.ContainsAny(suggestion.Value, "\r\n") || len(suggestion.Evidence) == 0 {
			return errors.New("bootstrap suggestion is invalid")
		}
		if suggestion.Confidence.Score < 0 || suggestion.Confidence.Score > 1 || suggestion.Confidence.EvidenceCount < 1 {
			return errors.New("bootstrap suggestion confidence is invalid")
		}
		for _, evidence := range suggestion.Evidence {
			if evidence.TenantID != p.TenantID || strings.TrimSpace(evidence.SourceType) == "" || strings.TrimSpace(evidence.SourceID) == "" {
				return errors.New("bootstrap suggestion evidence is outside tenant scope")
			}
		}
	}
	if p.CredentialsRequired && p.ReadyForReview {
		return errors.New("bootstrap plan cannot be ready while credentials are required")
	}
	return nil
}
