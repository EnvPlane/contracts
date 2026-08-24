package domain

import (
	"errors"
	"sort"
	"strings"
)

const AIRoutingSchemaVersion = "1"

type AIProviderCapability struct {
	SchemaVersion    string       `json:"schemaVersion"`
	ProviderID       string       `json:"providerId"`
	Mode             AIPolicyMode `json:"mode"`
	Model            string       `json:"model"`
	Region           string       `json:"region"`
	EndpointClass    string       `json:"endpointClass"`
	Capabilities     []string     `json:"capabilities"`
	SchemaVersionPin string       `json:"schemaVersionPin"`
	EvalGateVersion  string       `json:"evalGateVersion"`
	Priority         int          `json:"priority"`
}

type AIProviderRoutePolicy struct {
	SchemaVersion           string       `json:"schemaVersion"`
	Purpose                 string       `json:"purpose"`
	Mode                    AIPolicyMode `json:"mode"`
	AllowedRegions          []string     `json:"allowedRegions"`
	AllowedModels           []string     `json:"allowedModels"`
	RequiredCapability      string       `json:"requiredCapability"`
	PromptTemplateVersion   string       `json:"promptTemplateVersion"`
	SchemaVersionPin        string       `json:"schemaVersionPin"`
	RequiredEvalGateVersion string       `json:"requiredEvalGateVersion"`
	MaxFallbacks            int          `json:"maxFallbacks"`
}

type AIRouteRequest struct {
	SchemaVersion string `json:"schemaVersion"`
	TenantID      string `json:"tenantId"`
	Purpose       string `json:"purpose"`
	Capability    string `json:"capability"`
	Region        string `json:"region"`
}

type AIRouteDecision struct {
	SchemaVersion         string       `json:"schemaVersion"`
	ProviderID            string       `json:"providerId"`
	Mode                  AIPolicyMode `json:"mode"`
	Model                 string       `json:"model"`
	Region                string       `json:"region"`
	FallbackIndex         int          `json:"fallbackIndex"`
	Reason                string       `json:"reason"`
	PolicyVersion         string       `json:"policyVersion"`
	PromptTemplateVersion string       `json:"promptTemplateVersion"`
	SchemaVersionPin      string       `json:"schemaVersionPin"`
}

func (c AIProviderCapability) Validate() error {
	if c.SchemaVersion != AIRoutingSchemaVersion || strings.TrimSpace(c.ProviderID) == "" || strings.TrimSpace(c.Model) == "" || strings.TrimSpace(c.Region) == "" || strings.TrimSpace(c.EndpointClass) == "" || strings.TrimSpace(c.EvalGateVersion) == "" {
		return errors.New("AI provider capability identity is invalid")
	}
	if c.Mode != AIPolicyExternal && c.Mode != AIPolicySelfHosted {
		return errors.New("AI provider capability mode is invalid")
	}
	if len(c.Capabilities) == 0 || len(c.Capabilities) > 32 || c.Priority < 0 {
		return errors.New("AI provider capability is unbounded or invalid")
	}
	return nil
}

func (p AIProviderRoutePolicy) Validate() error {
	if p.SchemaVersion != AIRoutingSchemaVersion || strings.TrimSpace(p.Purpose) == "" || strings.TrimSpace(p.RequiredCapability) == "" || strings.TrimSpace(p.PromptTemplateVersion) == "" || strings.TrimSpace(p.SchemaVersionPin) == "" || strings.TrimSpace(p.RequiredEvalGateVersion) == "" || p.MaxFallbacks < 0 || p.MaxFallbacks > 3 {
		return errors.New("AI provider route policy is invalid")
	}
	if p.Mode != AIPolicyExternal && p.Mode != AIPolicySelfHosted {
		return errors.New("AI provider route policy mode is invalid")
	}
	if len(p.AllowedRegions) == 0 || len(p.AllowedRegions) > 32 || len(p.AllowedModels) == 0 || len(p.AllowedModels) > 64 {
		return errors.New("AI provider route policy allowlist is invalid")
	}
	return nil
}

func (r AIRouteRequest) Validate() error {
	if r.SchemaVersion != AIRoutingSchemaVersion || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.Purpose) == "" || strings.TrimSpace(r.Capability) == "" || strings.TrimSpace(r.Region) == "" {
		return errors.New("AI route request is invalid")
	}
	return nil
}

func (d AIRouteDecision) Validate() error {
	if d.SchemaVersion != AIRoutingSchemaVersion || strings.TrimSpace(d.ProviderID) == "" || strings.TrimSpace(d.Model) == "" || strings.TrimSpace(d.Region) == "" || strings.TrimSpace(d.Reason) == "" || strings.TrimSpace(d.PolicyVersion) == "" || strings.TrimSpace(d.PromptTemplateVersion) == "" || strings.TrimSpace(d.SchemaVersionPin) == "" || d.FallbackIndex < 0 {
		return errors.New("AI route decision is invalid")
	}
	if d.Mode != AIPolicyExternal && d.Mode != AIPolicySelfHosted {
		return errors.New("AI route decision mode is invalid")
	}
	return nil
}

func (p AIProviderRoutePolicy) Deterministic() AIProviderRoutePolicy {
	p.AllowedRegions = append([]string(nil), p.AllowedRegions...)
	p.AllowedModels = append([]string(nil), p.AllowedModels...)
	sort.Strings(p.AllowedRegions)
	sort.Strings(p.AllowedModels)
	return p
}
