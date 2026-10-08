package domain

import (
	"encoding/json"
	"testing"
)

func TestRemoteFinOpsBaselineOptIns(t *testing.T) {
	cfg := RemoteClusterFinOpsConfig{BaselineCapacityEnabled: true, BaselineMeasuredEnabled: true}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var persisted RemoteClusterFinOpsConfig
	if err := json.Unmarshal(b, &persisted); err != nil {
		t.Fatal(err)
	}
	if !persisted.BaselineCapacityEnabled || !persisted.BaselineMeasuredEnabled {
		t.Fatal("baseline opt-ins lost during persistence")
	}
	var defaults RemoteClusterFinOpsConfig
	if err := json.Unmarshal([]byte(`{}`), &defaults); err != nil {
		t.Fatal(err)
	}
	if defaults.BaselineCapacityEnabled || defaults.BaselineMeasuredEnabled {
		t.Fatal("baseline collection must remain opt-in")
	}
}

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
