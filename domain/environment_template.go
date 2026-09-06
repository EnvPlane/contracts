package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const EnvironmentTemplateContractVersion = "v1"

// ResourceTemplate contains sanitized desired state. Secret values are never
// represented here; secret references are explicit and non-materialized.
type ResourceTemplate struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Namespace  string         `json:"namespace,omitempty"`
	Name       string         `json:"name"`
	Manifest   map[string]any `json:"manifest"`
	Policy     string         `json:"policy,omitempty"`
}
type DependencyPolicy struct {
	Name            string                 `json:"name"`
	Mode            string                 `json:"mode"`
	Namespace       string                 `json:"namespace,omitempty"`
	Target          string                 `json:"target,omitempty"`
	Strategy        ResourcePolicyStrategy `json:"strategy,omitempty"`
	SourceNamespace string                 `json:"sourceNamespace,omitempty"`
	Required        bool                   `json:"required,omitempty"`
	Defaulted       bool                   `json:"defaulted,omitempty"`
	Reason          string                 `json:"reason,omitempty"`
	Path            string                 `json:"path,omitempty"`
}
type TemplateParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	SecretRef   string `json:"secretRef,omitempty"`
	Description string `json:"description,omitempty"`
}

type EnvironmentTemplateRevision struct {
	ContractVersion  string                     `json:"contractVersion"`
	RevisionID       string                     `json:"revisionId"`
	TemplateID       string                     `json:"templateId"`
	TenantID         string                     `json:"tenantId"`
	ProjectID        string                     `json:"projectId"`
	ClusterID        string                     `json:"clusterId"`
	SourceScanID     string                     `json:"sourceScanId"`
	SourceNamespaces []string                   `json:"sourceNamespaces"`
	ObservedVersions map[string]string          `json:"observedVersions,omitempty"`
	Resources        []ResourceTemplate         `json:"resources"`
	ResourcePolicies []ResourceDependencyPolicy `json:"resourcePolicies,omitempty"`
	Dependencies     []DependencyPolicy         `json:"dependencies,omitempty"`
	Parameters       []TemplateParameter        `json:"parameters,omitempty"`
	Digest           string                     `json:"digest"`
	CreatedAt        time.Time                  `json:"createdAt"`
	PublishedAt      time.Time                  `json:"publishedAt"`
}
type RenderInput struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
	Ref   string `json:"ref,omitempty"`
}
type RenderOutput struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Digest    string `json:"digest"`
}
type OwnershipRecord struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
type EnvironmentRenderInputs struct {
	TenantID             string            `json:"tenantId"`
	ProjectID            string            `json:"projectId"`
	EnvironmentID        string            `json:"environmentId"`
	DisplayName          string            `json:"displayName,omitempty"`
	Branch               string            `json:"branch,omitempty"`
	BranchSlug           string            `json:"branchSlug,omitempty"`
	ChangeID             string            `json:"changeId,omitempty"`
	CommitSHA            string            `json:"commitSha,omitempty"`
	ComponentImages      map[string]string `json:"componentImages,omitempty"`
	NamespaceMap         map[string]string `json:"namespaceMap"`
	ResourceNames        map[string]string `json:"resourceNames,omitempty"`
	Hostnames            map[string]string `json:"hostnames,omitempty"`
	ProjectDomain        string            `json:"projectDomain,omitempty"`
	PreviewDomainPattern string            `json:"previewDomainPattern,omitempty"`
	Backend              DeploymentBackend `json:"backend"`
	ImmutableImages      bool              `json:"immutableImages"`
}
type RenderedResource struct {
	ResourceID string         `json:"resourceId"`
	Kind       string         `json:"kind"`
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace,omitempty"`
	Manifest   map[string]any `json:"manifest"`
	Digest     string         `json:"digest"`
}
type TransformationReport struct {
	ResourceID string `json:"resourceId"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	FromDigest string `json:"fromDigest"`
	ToDigest   string `json:"toDigest"`
	Reason     string `json:"reason"`
}
type EnvironmentReleasePlan struct {
	ContractVersion                 string                 `json:"contractVersion"`
	PlanID                          string                 `json:"planId"`
	TenantID                        string                 `json:"tenantId"`
	ProjectID                       string                 `json:"projectId"`
	EnvironmentID                   string                 `json:"environmentId"`
	TemplateRevisionID              string                 `json:"templateRevisionId"`
	TemplateDigest                  string                 `json:"templateDigest"`
	Backend                         DeploymentBackend      `json:"backend"`
	Inputs                          []RenderInput          `json:"inputs,omitempty"`
	Outputs                         []RenderOutput         `json:"outputs,omitempty"`
	Ownership                       []OwnershipRecord      `json:"ownership"`
	RenderedResources               []RenderedResource     `json:"renderedResources"`
	Transformations                 []TransformationReport `json:"transformations"`
	Endpoints                       []IngressEndpoint      `json:"endpoints,omitempty"`
	PrimaryEndpoint                 string                 `json:"primaryEndpoint,omitempty"`
	SecretMaterializationPlanID     string                 `json:"secretMaterializationPlanId,omitempty"`
	SecretMaterializationPlanDigest string                 `json:"secretMaterializationPlanDigest,omitempty"`
	StatefulExecutionPlanID         string                 `json:"statefulExecutionPlanId,omitempty"`
	StatefulExecutionPlanDigest     string                 `json:"statefulExecutionPlanDigest,omitempty"`
	InputDigest                     string                 `json:"inputDigest"`
	ExecutionInputDigest            string                 `json:"executionInputDigest,omitempty"`
	Digest                          string                 `json:"digest"`
	CreatedAt                       time.Time              `json:"createdAt"`
}

// ReleasePlanDeploymentEnvironment is the stable Environment projection used
// by deployment backends. It intentionally excludes operational observations
// such as status, timestamps, cost estimates, endpoints, expiry, and pinning;
// those values can change without changing the desired deployment.
type ReleasePlanDeploymentEnvironment struct {
	TenantID                  string                 `json:"tenant_id,omitempty"`
	ID                        string                 `json:"id"`
	Project                   string                 `json:"project"`
	Product                   string                 `json:"product"`
	ClusterID                 string                 `json:"clusterId,omitempty"`
	Namespace                 string                 `json:"namespace"`
	TargetNamespace           string                 `json:"targetNamespace,omitempty"`
	HelmReleaseName           string                 `json:"helmReleaseName,omitempty"`
	Mode                      EnvironmentMode        `json:"mode"`
	Domain                    string                 `json:"domain"`
	Source                    SCMSource              `json:"source"`
	Base                      BaseEnvironment        `json:"base"`
	GitOps                    GitOpsTarget           `json:"gitops"`
	Charts                    ChartVersions          `json:"charts"`
	Infrastructure            Infrastructure         `json:"infrastructure"`
	Services                  []ServiceOverride      `json:"services"`
	Components                []EnvironmentComponent `json:"components,omitempty"`
	Overrides                 map[string]string      `json:"overrides,omitempty"`
	TTLHours                  int                    `json:"ttlHours"`
	TemplateRevisionID        string                 `json:"templateRevisionId,omitempty"`
	TemplateDigest            string                 `json:"templateDigest,omitempty"`
	DesiredRevision           EnvironmentRevision    `json:"desiredRevision"`
	ManifestPath              string                 `json:"manifestPath"`
	NamespaceManifestPath     string                 `json:"namespaceManifestPath,omitempty"`
	KustomizationManifestPath string                 `json:"kustomizationManifestPath,omitempty"`
}

func releasePlanDeploymentEnvironment(environment Environment) ReleasePlanDeploymentEnvironment {
	return ReleasePlanDeploymentEnvironment{
		TenantID: environment.TenantID, ID: environment.ID, Project: environment.Project,
		Product: environment.Product, ClusterID: environment.ClusterID, Namespace: environment.Namespace,
		TargetNamespace: environment.TargetNamespace, HelmReleaseName: environment.HelmReleaseName,
		Mode: environment.Mode, Domain: environment.Domain, Source: environment.Source,
		Base: environment.Base, GitOps: environment.GitOps, Charts: environment.Charts,
		Infrastructure: environment.Infrastructure, Services: environment.Services,
		Components: environment.Components, Overrides: environment.Overrides, TTLHours: environment.TTLHours,
		TemplateRevisionID: environment.TemplateRevisionID, TemplateDigest: environment.TemplateDigest,
		DesiredRevision: environment.DesiredRevision, ManifestPath: environment.ManifestPath,
		NamespaceManifestPath: environment.NamespaceManifestPath, KustomizationManifestPath: environment.KustomizationManifestPath,
	}
}

// ReleasePlanExecutionInputs identifies every command input consumed by a deployment backend.
type ReleasePlanExecutionInputs struct {
	Environment          ReleasePlanDeploymentEnvironment `json:"environment"`
	ProjectConfig        ProjectConfig                    `json:"projectConfig"`
	ChartRef             string                           `json:"chartRef,omitempty"`
	ChartVersion         string                           `json:"chartVersion,omitempty"`
	ProjectConfigVersion int                              `json:"projectConfigVersion,omitempty"`
}

// ReleasePlanExecutionInputDigest returns the digest that binds backend execution inputs to a signed plan.
func ReleasePlanExecutionInputDigest(environment Environment, projectConfig ProjectConfig, chartRef, chartVersion string, projectConfigVersion int) (string, error) {
	payload, err := json.Marshal(ReleasePlanExecutionInputs{
		Environment:          releasePlanDeploymentEnvironment(environment),
		ProjectConfig:        projectConfig,
		ChartRef:             chartRef,
		ChartVersion:         chartVersion,
		ProjectConfigVersion: projectConfigVersion,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func canonicalDigest(value any, clear func(any)) (string, error) {
	clear(value)
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
func (r EnvironmentTemplateRevision) CanonicalDigest() (string, error) {
	c := r
	return canonicalDigest(&c, func(v any) { v.(*EnvironmentTemplateRevision).Digest = "" })
}
func (p EnvironmentReleasePlan) CanonicalDigest() (string, error) {
	c := p
	return canonicalDigest(&c, func(v any) { v.(*EnvironmentReleasePlan).Digest = "" })
}
func (p EnvironmentReleasePlan) Validate() error {
	if p.ContractVersion != EnvironmentTemplateContractVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.EnvironmentID) == "" || strings.TrimSpace(p.TemplateRevisionID) == "" || strings.TrimSpace(p.TemplateDigest) == "" || strings.TrimSpace(p.InputDigest) == "" || strings.TrimSpace(p.Digest) == "" {
		return errors.New("invalid environment release plan binding")
	}
	actual, err := p.CanonicalDigest()
	if err != nil {
		return err
	}
	if actual != p.Digest {
		return fmt.Errorf("release plan digest mismatch: got %s want %s", p.Digest, actual)
	}
	return nil
}
func (r EnvironmentTemplateRevision) Validate() error {
	if r.ContractVersion != EnvironmentTemplateContractVersion || strings.TrimSpace(r.RevisionID) == "" || strings.TrimSpace(r.TemplateID) == "" || strings.TrimSpace(r.TenantID) == "" || strings.TrimSpace(r.ProjectID) == "" || strings.TrimSpace(r.SourceScanID) == "" {
		return errors.New("invalid environment template revision identity")
	}
	for _, resource := range r.Resources {
		if strings.EqualFold(resource.Kind, "Secret") {
			return errors.New("secret resources and secret bytes are not allowed in a template revision")
		}
	}
	for _, policy := range r.ResourcePolicies {
		if policy.Defaulted && policy.Strategy == ResourcePolicyUnsupported {
			return fmt.Errorf("explicit strategy required for %s", policy.ResourceID)
		}
	}
	if r.Digest == "" {
		return errors.New("template revision digest is required")
	}
	actual, err := r.CanonicalDigest()
	if err != nil {
		return err
	}
	if actual != r.Digest {
		return fmt.Errorf("template revision digest mismatch: got %s want %s", r.Digest, actual)
	}
	return nil
}
