package domain

import (
	"testing"
	"time"
)

func TestNetworkPolicyProbeReportValidatesSafeEvidence(t *testing.T) {
	valid := NetworkPolicyProbeReport{SchemaVersion: 1, ClusterUID: "cluster-uid", Generation: 1, CheckedAt: time.Now(), Scope: "single-node-pod-ipv4-tcp", Ingress: "passed", Egress: "passed", State: "passed", Reason: "measured", CleanupComplete: true}
	if valid.Validate() != nil {
		t.Fatal("valid report rejected")
	}
	for _, change := range []func(*NetworkPolicyProbeReport){
		func(r *NetworkPolicyProbeReport) { r.Generation = 0 },
		func(r *NetworkPolicyProbeReport) { r.Scope = "all-traffic" },
		func(r *NetworkPolicyProbeReport) { r.CleanupComplete = false },
		func(r *NetworkPolicyProbeReport) { r.Egress = "unknown" },
		func(r *NetworkPolicyProbeReport) { r.Reason = "raw-secret-output" },
	} {
		r := valid
		change(&r)
		if r.Validate() == nil {
			t.Fatalf("invalid report accepted: %+v", r)
		}
	}
}
