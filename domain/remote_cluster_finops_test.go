package domain

import "testing"

func TestRemoteFinOpsOriginValidation(t *testing.T) {
	for _, endpoint := range []string{"http://metrics.local", "https://token@metrics.local", "https://metrics.local?q=secret", "https://other.local"} {
		cfg := RemoteClusterFinOpsConfig{PrometheusEndpoint: endpoint, AllowedOrigins: []string{"https://metrics.local"}}
		if cfg.Validate() == nil {
			t.Fatalf("accepted unsafe endpoint %q", endpoint)
		}
	}
	cfg := RemoteClusterFinOpsConfig{PrometheusEndpoint: "https://metrics.local", AllowedOrigins: []string{"https://metrics.local"}, StorageUsedMetric: "envplane_pvc_directory_allocated_bytes"}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	cfg.StorageUsedMetric = "node_filesystem_used_bytes"
	if cfg.Validate() == nil {
		t.Fatal("accepted node-wide data as PVC metric")
	}
}
