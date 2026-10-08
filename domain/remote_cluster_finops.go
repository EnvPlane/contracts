package domain

import (
	"errors"
	"net/url"
	"strings"
)

// RemoteClusterFinOpsConfig persists opt-in collection, not metrics credentials
// or raw CA contents. The referenced CA must already exist in each Agent's
// target namespace; the installer must not copy credentials across projects.
type RemoteClusterFinOpsConfig struct {
	NodeInventoryEnabled         bool             `json:"node_inventory_enabled,omitempty"`
	CadvisorContainerdUIDEnabled bool             `json:"cadvisor_containerd_uid_enabled,omitempty"`
	PrometheusEndpoint           string           `json:"prometheus_endpoint,omitempty"`
	AllowedOrigins               []string         `json:"allowed_origins,omitempty"`
	TLS                          RemoteClusterTLS `json:"tls,omitempty"`
	StorageUsedMetric            string           `json:"storage_used_metric,omitempty"`
}

func (c *RemoteClusterFinOpsConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.StorageUsedMetric != "" && c.StorageUsedMetric != "kubelet_volume_stats_used_bytes" && c.StorageUsedMetric != "envplane_pvc_directory_allocated_bytes" {
		return errors.New("unapproved storage used metric")
	}
	if c.PrometheusEndpoint == "" {
		if len(c.AllowedOrigins) != 0 || c.TLS.CASecretRef != nil || c.TLS.ServerName != "" || c.CadvisorContainerdUIDEnabled {
			return errors.New("metrics options require a Prometheus endpoint")
		}
		return nil
	}
	u, err := url.Parse(c.PrometheusEndpoint)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(c.PrometheusEndpoint, "\r\n") {
		return errors.New("Prometheus endpoint must be credential-free HTTPS")
	}
	if len(c.AllowedOrigins) == 0 || len(c.AllowedOrigins) > 8 {
		return errors.New("bounded exact metrics origin allowlist required")
	}
	matched := false
	for _, origin := range c.AllowedOrigins {
		allowed, e := url.Parse(origin)
		if e != nil || allowed.Scheme != "https" || allowed.Hostname() == "" || allowed.User != nil || allowed.RawQuery != "" || allowed.Fragment != "" || (allowed.Path != "" && allowed.Path != "/") {
			return errors.New("metrics origins must be exact HTTPS origins")
		}
		if allowed.Scheme == u.Scheme && allowed.Host == u.Host {
			matched = true
		}
	}
	if !matched {
		return errors.New("Prometheus endpoint origin is not allowed")
	}
	if c.TLS.CASecretRef != nil && (c.TLS.CASecretRef.Name == "" || c.TLS.CASecretRef.Key == "") {
		return errors.New("metrics CA Secret name and key required")
	}
	return nil
}
