package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCompileStatefulExecutionPlanSupportsStrategiesWithoutSourceData(t *testing.T) {
	policies := []StatefulDependencyPolicy{
		{ID: "empty", Kind: "PersistentVolumeClaim", Strategy: StatefulStrategyEmpty, TargetName: "empty", TargetNamespace: "feature-a", Size: "5Gi", AccessModes: []string{"ReadWriteOnce"}},
		{ID: "seed", Kind: "Database", Strategy: StatefulStrategySeed, TargetName: "seed", TargetNamespace: "feature-a", SeedTemplateRef: "seed-v1", ServiceName: "seed-db", SecretRef: "seed-secret"},
		{ID: "snapshot", Kind: "PersistentVolumeClaim", Strategy: StatefulStrategyVolumeSnapshot, SourceNamespace: "base", SourceName: "data", TargetName: "clone", TargetNamespace: "feature-a", StorageClass: "fast", SnapshotClass: "csi-snap", Size: "10Gi", AccessModes: []string{"ReadWriteOnce"}, CSIProvisioner: "csi.example"},
		{ID: "restore", Kind: "Database", Strategy: StatefulStrategyDatabaseRestore, TargetName: "restore", TargetNamespace: "feature-a", DumpRef: "dump/object/ref", RestoreCredentialRef: "secret/restore", ServiceName: "restore-db", SecretRef: "restore-secret"},
	}
	plan, err := CompileStatefulExecutionPlan("tenant-a", "project-a", "env-a", "rev-1", "sha256:revision", "feature-a", policies, "sha256:input", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.Validate(); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(plan)
	text := strings.ToLower(string(payload))
	for _, forbidden := range []string{"dump bytes", "password", "source data", "plaintext"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("unsafe stateful payload contains %q", forbidden)
		}
	}
	if plan.CanDeleteSource("PersistentVolumeClaim", "base", "data") {
		t.Fatal("source PVC must never be cleanup eligible")
	}
	if err := EnforceStatefulQuota(plan, "1Gi", true); err == nil {
		t.Fatal("stateful storage quota must reject oversized clone")
	}
	if len(plan.DSNRewrites) != 2 || plan.DSNRewrites[0].TargetNamespace != "feature-a" {
		t.Fatalf("DSN rewrites must target feature-owned service/namespace: %+v", plan.DSNRewrites)
	}
	if len(plan.Readiness) != 4 {
		t.Fatalf("readiness gates = %d", len(plan.Readiness))
	}
}

func TestStatefulPlanBlocksMissingRequiredStrategyAndCapabilities(t *testing.T) {
	if _, err := CompileStatefulExecutionPlan("tenant-a", "project-a", "env-a", "rev-1", "sha256:revision", "feature-a", []StatefulDependencyPolicy{{ID: "db", Kind: "Database", Required: true, TargetName: "db", TargetNamespace: "feature-a"}}, "sha256:input", time.Now()); err == nil {
		t.Fatal("required dependency without strategy must block compilation")
	}
	if _, err := CompileStatefulExecutionPlan("tenant-a", "project-a", "env-a", "rev-1", "sha256:revision", "feature-a", []StatefulDependencyPolicy{{ID: "snap", Kind: "PersistentVolumeClaim", Strategy: StatefulStrategyVolumeSnapshot, SourceNamespace: "base", SourceName: "data", TargetName: "clone", TargetNamespace: "feature-a"}}, "sha256:input", time.Now()); err == nil {
		t.Fatal("snapshot without CSI/storage capabilities must block compilation")
	}
}
