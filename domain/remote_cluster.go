package domain

import "time"

// RemoteCluster is the safe desired-state record for a Kubernetes cluster
// managed from the EnvPlane control plane. Secrets are always references to
// operator-managed Kubernetes Secrets; raw kubeconfig, token and CA data are
// intentionally not part of this API model.
type RemoteCluster struct {
	TenantID          string                          `json:"tenant_id,omitempty"`
	ID                string                          `json:"id"`
	Name              string                          `json:"name"`
	Ownership         RemoteClusterOwnership          `json:"ownership"`
	Kubernetes        RemoteClusterKubernetesConfig   `json:"kubernetes"`
	ControlPlane      RemoteClusterControlPlaneConfig `json:"control_plane"`
	Agent             RemoteClusterAgentConfig        `json:"agent"`
	Runner            RemoteClusterRunnerConfig       `json:"runner"`
	Discovery         RemoteClusterDiscoveryScope     `json:"discovery"`
	FeatureNamespaces FeatureNamespacePolicy          `json:"feature_namespaces"`
	Status            RemoteClusterStatus             `json:"status,omitempty"`
	CreatedAt         time.Time                       `json:"created_at"`
	UpdatedAt         time.Time                       `json:"updated_at"`
}

type RemoteClusterOwnership struct {
	AccessUsers         []string `json:"access_users,omitempty"`
	AccessOrganizations []string `json:"access_organizations,omitempty"`
}

type SecretKeyReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key"`
}

type RemoteClusterKubernetesConfig struct {
	Endpoint                    string             `json:"endpoint"`
	CredentialSecretRef         SecretKeyReference `json:"credential_secret_ref"`
	CredentialSecretFingerprint string             `json:"credential_secret_fingerprint,omitempty"`
	TLS                         RemoteClusterTLS   `json:"tls,omitempty"`
}

type RemoteClusterTLS struct {
	ServerName  string              `json:"server_name,omitempty"`
	CASecretRef *SecretKeyReference `json:"ca_secret_ref,omitempty"`
}

type RemoteClusterControlPlaneConfig struct {
	Endpoint string                         `json:"endpoint"`
	TLS      RemoteClusterTLS               `json:"tls,omitempty"`
	Trust    RemoteClusterTrustDistribution `json:"trust,omitempty"`
}

// RemoteClusterTrustDistribution configures only the ownership and safe
// references for private CA trust. The raw CA is never persisted in a
// RemoteCluster record or returned by the API.
type RemoteClusterTrustDistribution struct {
	// Mode is existing (an operator-managed target Secret) or managementCopy
	// (copy the authorized management endpoint CA Secret into the target).
	Mode            string              `json:"mode,omitempty"`
	SourceSecretRef *SecretKeyReference `json:"source_secret_ref,omitempty"`
	TargetSecretRef *SecretKeyReference `json:"target_secret_ref,omitempty"`
}

type RemoteClusterTrustStatus struct {
	Mode            string              `json:"mode,omitempty"`
	SourceSecretRef *SecretKeyReference `json:"source_secret_ref,omitempty"`
	TargetSecretRef *SecretKeyReference `json:"target_secret_ref,omitempty"`
	Fingerprint     string              `json:"fingerprint,omitempty"`
	Revision        string              `json:"revision,omitempty"`
	ObservedAt      *time.Time          `json:"observed_at,omitempty"`
}

type RemoteClusterAgentConfig struct {
	Enabled            bool                `json:"enabled"`
	ReleaseName        string              `json:"release_name,omitempty"`
	Namespace          string              `json:"namespace,omitempty"`
	ChartVersion       string              `json:"chart_version,omitempty"`
	ImageReference     string              `json:"image_reference,omitempty"`
	ImagePullSecretRef *SecretKeyReference `json:"image_pull_secret_ref,omitempty"`
	ProjectID          string              `json:"project_id,omitempty"`
}

type RemoteClusterRunnerConfig struct {
	Enabled            bool                `json:"enabled"`
	ReleaseName        string              `json:"release_name,omitempty"`
	Namespace          string              `json:"namespace,omitempty"`
	ChartVersion       string              `json:"chart_version,omitempty"`
	ImageReference     string              `json:"image_reference,omitempty"`
	ImagePullSecretRef *SecretKeyReference `json:"image_pull_secret_ref,omitempty"`
	ProjectID          string              `json:"project_id,omitempty"`
}

type RemoteClusterDiscoveryScope struct {
	NamespaceSelector  string   `json:"namespace_selector,omitempty"`
	AllowedNamespaces  []string `json:"allowed_namespaces,omitempty"`
	ExcludedNamespaces []string `json:"excluded_namespaces,omitempty"`
}

type FeatureNamespacePolicy struct {
	Mode            string   `json:"mode"`
	SharedNamespace string   `json:"shared_namespace,omitempty"`
	AllowedPrefixes []string `json:"allowed_prefixes,omitempty"`
	MaxNamespaces   int      `json:"max_namespaces,omitempty"`
}

// RemoteClusterCondition is an observed, safe reconciliation fact. Conditions
// deliberately describe only operator actions and never include kubeconfig,
// bootstrap-token, Helm values, or remote API response content.
type RemoteClusterCondition struct {
	Type               string     `json:"type"`
	Status             string     `json:"status"`
	Reason             string     `json:"reason,omitempty"`
	ObservedGeneration int64      `json:"observed_generation,omitempty"`
	LastTransitionAt   *time.Time `json:"last_transition_at,omitempty"`
}

// ManagementEndpointPreflight is the deliberately small observation reported
// by a managed Agent or Runner after it probes the management endpoint from
// its own target-cluster Pod.  It never contains an endpoint URL, certificate
// material, HTTP response body, or authorization data.
//
// Code is an allow-listed diagnostic (passed, dns_failed, tcp_failed,
// tls_ca_failed, tls_server_name_mismatch, endpoint_unhealthy or
// runtime_auth_failed).  The control plane validates the source identity and
// records the authenticated runtime-access result itself.
type ManagementEndpointPreflight struct {
	Generation      int64      `json:"generation"`
	Code            string     `json:"code"`
	DNSResolved     bool       `json:"dns_resolved"`
	TCPConnected    bool       `json:"tcp_connected"`
	TLSVerified     bool       `json:"tls_verified"`
	HealthReachable bool       `json:"health_reachable"`
	RuntimeAccess   bool       `json:"runtime_access"`
	CheckedAt       *time.Time `json:"checked_at,omitempty"`
}

// RemoteClusterAttempt identifies one reconciliation pass. It is persisted so
// a UI/API client can distinguish an in-flight repair from a historical
// success without seeing credentials or a raw executor error.
type RemoteClusterAttempt struct {
	ID          string     `json:"id"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Action      string     `json:"action,omitempty"`
}

// RemoteClusterArtifactRef records immutable artifacts actually selected by
// the signed active umbrella compatibility manifest.
type RemoteClusterArtifactRef struct {
	Component                string `json:"component"`
	ReleaseName              string `json:"release_name"`
	Namespace                string `json:"namespace"`
	ChartRef                 string `json:"chart_ref"`
	ChartVersion             string `json:"chart_version"`
	Image                    string `json:"image"`
	CompatibilityFingerprint string `json:"compatibility_fingerprint,omitempty"`
	Generation               int64  `json:"generation,omitempty"`
}

// RemoteClusterMigration is an explicit, audited request to take over a
// legacy Agent/Runner release. The reconciler never adopts an unmanaged
// release merely because its name happens to match the desired state.
type RemoteClusterMigration struct {
	RequestedAt *time.Time `json:"requested_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

// RemoteClusterRemoval tracks a controlled remote cleanup. It intentionally
// never includes raw Helm/Kubernetes errors or credential material.
type RemoteClusterRemoval struct {
	RequestedAt *time.Time `json:"requested_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

// RemoteClusterStatus is durable reconciler observation. DesiredGeneration is
// advanced only by desired-state/repair changes; ObservedGeneration advances
// only after a completed pass. Generation remains as a backwards-compatible
// alias for DesiredGeneration.
type RemoteClusterStatus struct {
	Phase              string `json:"phase,omitempty"` // installing|healthy|stale|degraded|blocked|removing
	Reason             string `json:"reason,omitempty"`
	DesiredGeneration  int64  `json:"desired_generation,omitempty"`
	ObservedGeneration int64  `json:"observed_generation,omitempty"`
	// ManagementEndpointProfile generations bind the target rollout and
	// preflight to the persisted endpoint/TLS contract. They are safe numbers,
	// never endpoint credentials or certificate material.
	ManagementEndpointProfileDesiredGeneration  int64                    `json:"management_endpoint_profile_desired_generation,omitempty"`
	ManagementEndpointProfileObservedGeneration int64                    `json:"management_endpoint_profile_observed_generation,omitempty"`
	Generation                                  int64                    `json:"generation,omitempty"`
	ObservedAt                                  *time.Time               `json:"observed_at,omitempty"`
	LastAttemptAt                               *time.Time               `json:"last_attempt_at,omitempty"`
	LastHealthyAt                               *time.Time               `json:"last_healthy_at,omitempty"`
	Attempt                                     *RemoteClusterAttempt    `json:"attempt,omitempty"`
	AttemptID                                   string                   `json:"attempt_id,omitempty"`
	RetryAt                                     *time.Time               `json:"retry_at,omitempty"`
	Conditions                                  []RemoteClusterCondition `json:"conditions,omitempty"`
	// EndpointPreflight contains only safe target-Pod observations for the
	// current desired generation.  It is cleared whenever desired state changes.
	EndpointPreflight  *RemoteClusterEndpointPreflightStatus `json:"endpoint_preflight,omitempty"`
	Trust              *RemoteClusterTrustStatus             `json:"trust,omitempty"`
	InstalledArtifacts []RemoteClusterArtifactRef            `json:"installed_artifacts,omitempty"`
	// RunnerHelmTargetNamespaces is controller-owned desired state. It contains
	// only concrete feature namespaces that have passed the RemoteCluster
	// namespace policy; it never represents a wildcard/prefix grant.
	RunnerHelmTargetNamespaces []string                `json:"runner_helm_target_namespaces,omitempty"`
	RecoveryAction             string                  `json:"recovery_action,omitempty"` // retry|rollback|rotate|repair|migrate|remove
	Migration                  *RemoteClusterMigration `json:"migration,omitempty"`
	Removal                    *RemoteClusterRemoval   `json:"removal,omitempty"`
}

type RemoteClusterEndpointPreflightStatus struct {
	Generation int64                        `json:"generation"`
	Agent      *ManagementEndpointPreflight `json:"agent,omitempty"`
	Runner     *ManagementEndpointPreflight `json:"runner,omitempty"`
}
