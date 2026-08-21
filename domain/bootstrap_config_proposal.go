package domain

import (
	"errors"
	"sort"
	"strings"
)

const BootstrapConfigProposalSchemaVersion = "1"

type BootstrapConfigProposalRequest struct {
	Intent string `json:"intent"`
}

type BootstrapPolicyDefaults struct {
	DefaultTTLHours         *int `json:"defaultTTLHours,omitempty"`
	MaxActiveEnvironments   *int `json:"maxActiveEnvironments,omitempty"`
	MaxCPUPerEnvironment    *int `json:"maxCPUPerEnvironment,omitempty"`
	MaxMemoryPerEnvironment *int `json:"maxMemoryPerEnvironment,omitempty"`
}

type BootstrapConfigProposalFields struct {
	Mode                   *EnvironmentMode            `json:"mode,omitempty"`
	ComponentRefs          []ConfigurationComponentRef `json:"componentRefs,omitempty"`
	GitOpsRepositoryTarget *RepositoryRef              `json:"gitOpsRepositoryTarget,omitempty"`
	SecretStrategies       map[string]string           `json:"secretStrategies,omitempty"`
	TTLHours               *int                        `json:"ttlHours,omitempty"`
	PolicyDefaults         *BootstrapPolicyDefaults    `json:"policyDefaults,omitempty"`
}

type BootstrapConfigProposalDiff struct {
	Field  string `json:"field"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

type BootstrapConfigProposalRejectedField struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type BootstrapConfigProposal struct {
	SchemaVersion      string                                 `json:"schemaVersion"`
	ProposalVersion    string                                 `json:"proposalVersion"`
	Kind               string                                 `json:"kind"`
	TenantID           string                                 `json:"tenantId"`
	ProjectID          string                                 `json:"projectId"`
	Fields             BootstrapConfigProposalFields          `json:"fields"`
	Diff               []BootstrapConfigProposalDiff          `json:"diff,omitempty"`
	Warnings           []string                               `json:"warnings,omitempty"`
	RejectedFields     []BootstrapConfigProposalRejectedField `json:"rejectedFields,omitempty"`
	Valid              bool                                   `json:"valid"`
	ContextHash        string                                 `json:"contextHash"`
	RepoProfileVersion string                                 `json:"repoProfileVersion"`
}

func (f BootstrapConfigProposalFields) Deterministic() BootstrapConfigProposalFields {
	f.ComponentRefs = append([]ConfigurationComponentRef(nil), f.ComponentRefs...)
	sort.Slice(f.ComponentRefs, func(i, j int) bool { return f.ComponentRefs[i].ComponentID < f.ComponentRefs[j].ComponentID })
	if f.SecretStrategies != nil {
		copyMap := make(map[string]string, len(f.SecretStrategies))
		for key, value := range f.SecretStrategies {
			copyMap[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
		f.SecretStrategies = copyMap
	}
	return f
}

func (f BootstrapConfigProposalFields) Validate() error {
	if f.Mode != nil && *f.Mode != ModeFull && *f.Mode != ModeHybrid {
		return errors.New("unsupported bootstrap mode")
	}
	if f.TTLHours != nil && (*f.TTLHours < 1 || *f.TTLHours > 720) {
		return errors.New("bootstrap TTL must be between 1 and 720 hours")
	}
	seen := map[string]struct{}{}
	for _, ref := range f.ComponentRefs {
		id := strings.TrimSpace(ref.ComponentID)
		if id == "" {
			return errors.New("bootstrap component ID is required")
		}
		if _, ok := seen[id]; ok {
			return errors.New("duplicate bootstrap component ID")
		}
		seen[id] = struct{}{}
	}
	for key, strategy := range f.SecretStrategies {
		if strings.TrimSpace(key) == "" || (strategy != "" && strategy != "reference existing secret" && strategy != "external secret" && strategy != "encrypted clone" && strategy != "manual input") {
			return errors.New("unsupported secret strategy")
		}
	}
	if f.PolicyDefaults != nil {
		for _, value := range []*int{f.PolicyDefaults.DefaultTTLHours, f.PolicyDefaults.MaxActiveEnvironments, f.PolicyDefaults.MaxCPUPerEnvironment, f.PolicyDefaults.MaxMemoryPerEnvironment} {
			if value != nil && *value < 0 {
				return errors.New("bootstrap policy defaults cannot be negative")
			}
		}
	}
	return nil
}
