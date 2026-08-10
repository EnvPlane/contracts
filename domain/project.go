package domain

import "time"

type Project struct {
	TenantID              string             `json:"tenant_id,omitempty" db:"tenant_id"`
	ID                    string             `json:"id" db:"id"`
	Name                  string             `json:"name" db:"name"`
	ProductID             string             `json:"product_id" db:"product_id"`
	AppRepositoryID       string             `json:"app_repository_id,omitempty" db:"app_repository_id"`
	GitOpsRepositoryID    string             `json:"gitops_repository_id,omitempty" db:"gitops_repository_id"`
	WebhookBranchFilters  []string           `json:"branch_filters,omitempty" db:"branch_filters"`
	WebhookLabels         []string           `json:"labels,omitempty" db:"webhook_labels"`
	WebhookAllowDraftPRs  bool               `json:"allow_draft_prs,omitempty" db:"webhook_allow_draft_prs"`
	GitHubInstallationIDs []string           `json:"github_installation_ids,omitempty" db:"github_installation_ids"`
	GitLabProjectIDs      []string           `json:"gitlab_project_ids,omitempty" db:"gitlab_project_ids"`
	ClusterID             string             `json:"cluster_id,omitempty" db:"cluster_id"`
	AuthorizedClusterIDs  []string           `json:"authorized_cluster_ids,omitempty" db:"authorized_cluster_ids"`
	AccessUsers           []string           `json:"access_users,omitempty" db:"access_users"`
	AccessOrganizations   []string           `json:"access_organizations,omitempty" db:"access_organizations"`
	SecretRefs            []string           `json:"secret_refs,omitempty" db:"secret_refs"`
	GitRepo               RepositoryRef      `json:"git_repo" db:"git_repo"`
	GitOpsRepo            RepositoryRef      `json:"gitops_repo" db:"gitops_repo"`
	Components            []ProjectComponent `json:"components,omitempty" db:"components"`
	BaseEnvConfig         BaseEnvConfig      `json:"base_env_config" db:"base_env_config"`
	CostPolicy            ProjectCostPolicy  `json:"cost_policy,omitempty" db:"cost_policy"`
	CreatedAt             time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at" db:"updated_at"`
}

// ProjectDeploymentReadiness describes whether a project can be used to create
// a feature environment. It is derived from the bootstrap session and project
// topology; it is intentionally a response field rather than persisted project
// configuration.
type ProjectDeploymentReadiness struct {
	Ready                      bool                           `json:"ready"`
	BootstrapStatus            string                         `json:"bootstrap_status,omitempty"`
	ResourceScanStatus         string                         `json:"resource_scan_status,omitempty"`
	MissingPrerequisites       []string                       `json:"missing_prerequisites,omitempty"`
	NextAction                 *ProjectReadinessAction        `json:"next_action,omitempty"`
	BaseNamespaces             []string                       `json:"base_namespaces,omitempty"`
	BaseServices               []BaseServiceRef               `json:"base_services,omitempty"`
	HybridReady                bool                           `json:"hybrid_ready"`
	HybridMissingPrerequisites []string                       `json:"hybrid_missing_prerequisites,omitempty"`
	RemoteCluster              *ProjectRemoteClusterReadiness `json:"remote_cluster,omitempty"`
}

// ProjectRemoteClusterReadiness exposes only safe observed state for the
// selected project target. It lets clients distinguish a bootstrap gap from a
// stale/degraded remote execution target before Environment creation is tried.
type ProjectRemoteClusterReadiness struct {
	ID                                          string     `json:"id"`
	Phase                                       string     `json:"phase"`
	Reason                                      string     `json:"reason,omitempty"`
	Fresh                                       bool       `json:"fresh"`
	RecoveryAction                              string     `json:"recovery_action,omitempty"`
	ObservedAt                                  *time.Time `json:"observed_at,omitempty"`
	ManagementEndpointProfileDesiredGeneration  int64      `json:"management_endpoint_profile_desired_generation,omitempty"`
	ManagementEndpointProfileObservedGeneration int64      `json:"management_endpoint_profile_observed_generation,omitempty"`
	EndpointPreflightGeneration                 int64      `json:"endpoint_preflight_generation,omitempty"`
	AgentEndpointPreflightCode                  string     `json:"agent_endpoint_preflight_code,omitempty"`
	RunnerEndpointPreflightCode                 string     `json:"runner_endpoint_preflight_code,omitempty"`
	EndpointPreflightFresh                      bool       `json:"endpoint_preflight_fresh"`
}

// ProjectReadinessAction is a safe, server-derived pointer to the next
// bootstrap activity required before a project can create environments. It
// deliberately contains no credential material or bootstrap token data.
type ProjectReadinessAction struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Description     string `json:"description"`
	BootstrapStep   int    `json:"bootstrap_step"`
	RemoteClusterID string `json:"remote_cluster_id,omitempty"`
	RemoteAction    string `json:"remote_action,omitempty"`
	Href            string `json:"href,omitempty"`
}

// ProjectResponse is the public project representation. Embedding Project
// preserves the existing API contract while exposing deployment readiness to
// clients that create environments.
type ProjectResponse struct {
	Project
	DeploymentReadiness ProjectDeploymentReadiness `json:"deployment_readiness"`
}

type RepositoryRef struct {
	Provider      string `json:"provider"`
	URL           string `json:"url"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path,omitempty"`
}

// ProjectComponent describes one independently versioned application component.
// The service value is the name consumed by the deployment templates; it is kept
// separate from the repository so one repository can map to an arbitrary service.
type ProjectComponent struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	Service    string                  `json:"service"`
	Repository RepositoryRef           `json:"repository"`
	Build      ComponentBuildMetadata  `json:"build,omitempty"`
	Deploy     ComponentDeployMetadata `json:"deploy,omitempty"`
}

// ComponentBuildMetadata is deliberately provider-agnostic. It records the
// image identity and the tag to use when no per-environment override is given.
type ComponentBuildMetadata struct {
	Image      string `json:"image,omitempty"`
	DefaultTag string `json:"default_tag,omitempty"`
}

// ComponentDeployMetadata describes where the mapped service is rendered in
// GitOps without coupling the project record to a particular renderer.
type ComponentDeployMetadata struct {
	Chart      string `json:"chart,omitempty"`
	ValuesPath string `json:"values_path,omitempty"`
}

type BaseEnvConfig struct {
	EnvironmentID   string            `json:"environment_id"`
	Namespace       string            `json:"namespace"`
	Domain          string            `json:"domain"`
	ConfigPath      string            `json:"config_path"`
	Services        []BaseServiceRef  `json:"services,omitempty"`
	Values          map[string]string `json:"values,omitempty"`
	HybridOverrides map[string]bool   `json:"hybrid_overrides,omitempty"`
}

type BaseServiceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type ProjectCostPolicy struct {
	DefaultTTLHours         int   `json:"default_ttl_hours,omitempty"`
	MaxActiveEnvsPerProject int   `json:"max_active_envs_per_project,omitempty"`
	MaxCPUPerEnv            int   `json:"max_cpu_per_env,omitempty"`
	MaxMemoryPerEnv         int   `json:"max_memory_per_env,omitempty"`
	IdleTimeoutHours        int   `json:"idle_timeout_hours,omitempty"`
	AutoDeleteIdleEnvs      *bool `json:"auto_delete_idle_envs,omitempty"`
}
