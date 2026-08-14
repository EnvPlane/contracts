package domain

import "time"

type ControlPlaneSettings struct {
	Repositories    []ConfiguredRepository `json:"repositories"`
	SecretRefs      []SecretReference      `json:"secret_refs"`
	ManifestSources []ManifestSource       `json:"manifest_sources"`
	Clusters        []ClusterTarget        `json:"clusters"`
	Notifications   []NotificationTarget   `json:"notifications"`
	Runtime         RuntimeSettings        `json:"runtime"`
	// ManagementEndpointProfile is the single, API-managed endpoint contract
	// used by all remote clusters. It intentionally contains only safe URL and
	// Secret-reference metadata; certificate, kubeconfig and token bytes never
	// belong in settings or an API response.
	ManagementEndpointProfile *ManagementEndpointProfile `json:"management_endpoint_profile,omitempty"`
	Authentication            AuthenticationSettings     `json:"authentication"`
	AuthenticationTransaction *AuthenticationTransaction `json:"authentication_transaction,omitempty"`
	UpdatedAt                 time.Time                  `json:"updated_at"`
	UpdatedBy                 string                     `json:"updated_by,omitempty"`
	SchemaVersion             string                     `json:"schema_version"`
}

// AuthenticationTransaction is safe persisted coordination state. It never
// contains provider credentials or Secret references.
type AuthenticationTransaction struct {
	ID                         string    `json:"id"`
	Kind                       string    `json:"kind"`
	Status                     string    `json:"status"`
	ExpectedSettingsRevision   int64     `json:"expectedSettingsRevision"`
	ExpectedCredentialRevision int64     `json:"expectedCredentialRevision"`
	ExpectedRuntimeRevision    int64     `json:"expectedRuntimeRevision"`
	// ExpectedMaterialRevision fences an out-of-band setup recovery against the
	// exact managed-Secret material snapshot it inspected. It is intentionally
	// separate from runtime revisions, which belong to provider activation.
	ExpectedMaterialRevision string    `json:"expectedMaterialRevision,omitempty"`
	IdempotencyKey             string    `json:"idempotencyKey,omitempty"`
	ErrorCode                  string    `json:"errorCode,omitempty"`
	CreatedAt                  time.Time `json:"createdAt"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// ManagementEndpointProfile describes the management control-plane HTTPS
// address that remote Agent and Runner Pods use. It is deliberately separate
// from RemoteCluster: a remote target selects this canonical profile instead
// of storing an independently editable management endpoint.
type ManagementEndpointProfile struct {
	Endpoint           string                        `json:"endpoint"`
	TLS                RemoteClusterTLS              `json:"tls,omitempty"`
	Source             string                        `json:"source,omitempty"` // api|chart_bootstrap
	DesiredGeneration  int64                         `json:"desired_generation,omitempty"`
	ObservedGeneration int64                         `json:"observed_generation,omitempty"`
	ObservedAt         *time.Time                    `json:"observed_at,omitempty"`
	Conditions         []ManagementEndpointCondition `json:"conditions,omitempty"`
	CreatedAt          time.Time                     `json:"created_at,omitempty"`
	UpdatedAt          time.Time                     `json:"updated_at,omitempty"`
}

// ManagementEndpointCondition is a safe observed fact about the endpoint
// profile. Reasons are intentionally bounded identifiers, never TLS or
// network payloads that could disclose credential material.
type ManagementEndpointCondition struct {
	Type               string     `json:"type"`
	Status             string     `json:"status"`
	Reason             string     `json:"reason,omitempty"`
	ObservedGeneration int64      `json:"observed_generation,omitempty"`
	LastTransitionAt   *time.Time `json:"last_transition_at,omitempty"`
}

type ConfiguredRepository struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	Provider       string `json:"provider"`
	URL            string `json:"url"`
	DefaultBranch  string `json:"default_branch"`
	BranchStrategy string `json:"branch_strategy,omitempty"`
	Path           string `json:"path,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	SecretRef      string `json:"secret_ref,omitempty"`
}

type SecretReference struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Scope       string `json:"scope"`
	Reference   string `json:"reference"`
	Description string `json:"description,omitempty"`
}

type ManifestSource struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	RepositoryID string `json:"repository_id,omitempty"`
	Path         string `json:"path"`
	ValuesPath   string `json:"values_path,omitempty"`
	Version      string `json:"version,omitempty"`
	Enabled      bool   `json:"enabled"`
}

type ClusterTarget struct {
	ID                       string     `json:"id"`
	Name                     string     `json:"name"`
	Provider                 string     `json:"provider,omitempty"`
	APIURL                   string     `json:"api_url,omitempty"`
	Context                  string     `json:"context,omitempty"`
	AgentID                  string     `json:"agent_id,omitempty"`
	AgentNamespace           string     `json:"agent_namespace,omitempty"`
	AgentVersion             string     `json:"agent_version,omitempty"`
	AgentStatus              string     `json:"agent_status,omitempty"`
	AgentError               string     `json:"agent_error,omitempty"`
	NamespaceSelector        string     `json:"namespace_selector,omitempty"`
	SecretRef                string     `json:"secret_ref,omitempty"`
	KubernetesVersion        string     `json:"kubernetes_version,omitempty"`
	FluxNamespace            string     `json:"flux_namespace,omitempty"`
	Capabilities             []string   `json:"capabilities,omitempty"`
	HeartbeatIntervalSeconds int        `json:"heartbeat_interval_seconds,omitempty"`
	LastHeartbeatAt          *time.Time `json:"last_heartbeat_at,omitempty"`
	Enabled                  bool       `json:"enabled"`
}

type NotificationTarget struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Channel   string `json:"channel,omitempty"`
	SecretRef string `json:"secret_ref,omitempty"`
	Enabled   bool   `json:"enabled"`
}

type RuntimeSettings struct {
	DefaultProduct          string          `json:"default_product"`
	DefaultProject          string          `json:"default_project"`
	DefaultMode             EnvironmentMode `json:"default_mode"`
	DomainRoot              string          `json:"domain_root"`
	NamespacePrefix         string          `json:"namespace_prefix"`
	DefaultTTLHours         int             `json:"default_ttl_hours"`
	IdleThresholdHours      int             `json:"idle_threshold_hours,omitempty"`
	TTLCheckSeconds         int             `json:"ttl_check_seconds"`
	JobRetrySeconds         int             `json:"job_retry_seconds"`
	JobMaxAttempts          int             `json:"job_max_attempts"`
	MaxCPUPerEnv            int             `json:"max_cpu_per_env,omitempty"`
	MaxMemoryPerEnv         int             `json:"max_memory_per_env,omitempty"`
	MaxActiveEnvsPerProject int             `json:"max_active_envs_per_project"`
	AutoDeleteIdleEnvs      *bool           `json:"auto_delete_idle_envs,omitempty"`
	GitOpsDir               string          `json:"gitops_dir"`
	ProductBasePath         string          `json:"product_base_path"`
	FluxNamespace           string          `json:"flux_namespace"`
	SourceRefName           string          `json:"source_ref_name"`
	DependsOnName           string          `json:"depends_on_name"`
	HealthCheckName         string          `json:"health_check_name"`
	EnableGitCommit         bool            `json:"enable_git_commit"`
	EnableGitPush           bool            `json:"enable_git_push"`
	GitPushRemote           string          `json:"git_push_remote"`
	GitPushBranch           string          `json:"git_push_branch"`
	CatalogPath             string          `json:"catalog_path,omitempty"`
	DatabaseURLConfigured   bool            `json:"database_url_configured"`
	RedisURLConfigured      bool            `json:"redis_url_configured"`
}

type AgentRegistrationRequest struct {
	ProjectID                string                   `json:"projectId,omitempty"`
	ClusterID                string                   `json:"clusterId"`
	AgentID                  string                   `json:"agentId"`
	RegistrationToken        string                   `json:"registrationToken,omitempty"`
	AgentVersion             string                   `json:"agentVersion,omitempty"`
	AgentNamespace           string                   `json:"agentNamespace,omitempty"`
	KubernetesVersion        string                   `json:"kubernetesVersion,omitempty"`
	FluxNamespace            string                   `json:"fluxNamespace,omitempty"`
	NamespaceSelector        string                   `json:"namespaceSelector,omitempty"`
	Capabilities             []string                 `json:"capabilities,omitempty"`
	CapabilityReport         *ClusterCapabilityReport `json:"capabilityReport,omitempty"`
	HeartbeatIntervalSeconds int                      `json:"heartbeatIntervalSeconds,omitempty"`
	ObservedAt               time.Time                `json:"observedAt,omitempty"`
}

type AgentRegistrationResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Provider        string     `json:"provider"`
	AgentID         string     `json:"agent_id,omitempty"`
	AgentStatus     string     `json:"agent_status,omitempty"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	Enabled         bool       `json:"enabled"`
	AgentAuthToken  string     `json:"agentAuthToken,omitempty"`
}

type AgentHeartbeatRequest struct {
	ProjectID                string                       `json:"projectId,omitempty"`
	ClusterID                string                       `json:"clusterId"`
	AgentID                  string                       `json:"agentId"`
	AgentAuthToken           string                       `json:"agentAuthToken,omitempty"`
	AgentVersion             string                       `json:"agentVersion,omitempty"`
	KubernetesVersion        string                       `json:"kubernetesVersion,omitempty"`
	Capabilities             []string                     `json:"capabilities,omitempty"`
	CapabilityReport         *ClusterCapabilityReport     `json:"capabilityReport,omitempty"`
	HeartbeatIntervalSeconds int                          `json:"heartbeatIntervalSeconds,omitempty"`
	Status                   string                       `json:"status,omitempty"`
	Error                    string                       `json:"error,omitempty"`
	EndpointPreflight        *ManagementEndpointPreflight `json:"endpoint_preflight,omitempty"`
	ObservedAt               time.Time                    `json:"observedAt,omitempty"`
}
