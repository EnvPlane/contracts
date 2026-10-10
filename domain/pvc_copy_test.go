package domain

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func pvcCopyFixture(t *testing.T) (PVCCopyPlan, []PVCCopySource, PVCCopyPermission) {
	t.Helper()
	sources := []PVCCopySource{{TenantID: "tenant-a", Namespace: "base", Name: "data", UID: "uid-1", StorageClass: "standard", AccessModes: []string{"ReadWriteOnce"}, RequestedBytes: 1024, VolumeMode: "Filesystem", EnvironmentClass: "development"}}
	p := PVCCopyPlan{PlanID: "copy-1", TenantID: "tenant-a", ProjectID: "project-a", EnvironmentID: "preview-a", TemplateRevisionID: "revision-a", TemplateDigest: "sha256:" + strings.Repeat("a", 64), TargetNamespace: "preview-a", AllowedSourceNamespaces: []string{"base"}, Image: "registry.example/copy@sha256:" + strings.Repeat("b", 64), MaxBytes: 2048, TimeoutSeconds: 300, StorageQuotaBytes: 4096,
		Items: []PVCCopyItem{{ID: "data", SourceTenantID: "tenant-a", SourceNamespace: "base", SourceName: "data", SourceUID: "uid-1", SourceRequestedBytes: 1024, TargetName: "data-copy", StorageClass: "standard", AccessModes: []string{"ReadWriteOnce"}, RequestedBytes: 2048, Method: PVCCopyFilesystemOffline, MaxBytes: 2048, TimeoutSeconds: 300}}}
	permission := PVCCopyPermission{PlanID: p.PlanID, TenantID: p.TenantID, ProjectID: p.ProjectID, EnvironmentID: p.EnvironmentID, TemplateRevisionID: p.TemplateRevisionID, TemplateDigest: p.TemplateDigest, TargetNamespace: p.TargetNamespace, AllowedSourceNamespaces: []string{"base"}, AllowedImages: []string{p.Image}, MaxBytes: p.MaxBytes, TimeoutSeconds: p.TimeoutSeconds, StorageQuotaBytes: p.StorageQuotaBytes}
	return p, sources, permission
}

func TestPVCCopyCompileAndRoundTrip(t *testing.T) {
	p, sources, permission := pvcCopyFixture(t)
	original := p.Items[0]
	compiled, err := CompilePVCCopyPlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Items[0], original) || compiled.Items[0].SourceIdentity == "" || compiled.ContractVersion != PVCCopyContractVersion {
		t.Fatal("compiler mutated selection or failed to bind scan identity")
	}
	encoded, err := json.Marshal(compiled)
	if err != nil {
		t.Fatal(err)
	}
	var decoded PVCCopyPlan
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePVCCopyPlan(decoded, sources, permission); err != nil {
		t.Fatal(err)
	}
	compiled.Items[0].AccessModes[0] = "ReadWriteMany"
	compiled.AllowedSourceNamespaces[0] = "other"
	if p.Items[0].AccessModes[0] != "ReadWriteOnce" || p.AllowedSourceNamespaces[0] != "base" {
		t.Fatal("compiled plan aliases input slices")
	}
}

func TestPVCCopyRejectsInvalidSelection(t *testing.T) {
	tests := map[string]func(*PVCCopyPlan){
		"version":                   func(p *PVCCopyPlan) { p.ContractVersion = "v99" },
		"identity whitespace":       func(p *PVCCopyPlan) { p.TenantID = " tenant-a" },
		"uppercase digest":          func(p *PVCCopyPlan) { p.TemplateDigest = "sha256:" + strings.Repeat("A", 64) },
		"short digest":              func(p *PVCCopyPlan) { p.TemplateDigest = "sha256:abc" },
		"mutable image":             func(p *PVCCopyPlan) { p.Image = "registry.example/copy:latest" },
		"namespace shell":           func(p *PVCCopyPlan) { p.TargetNamespace = "a;id" },
		"target path":               func(p *PVCCopyPlan) { p.Items[0].TargetName = "../data" },
		"target shell":              func(p *PVCCopyPlan) { p.Items[0].TargetName = "data$(id)" },
		"source path":               func(p *PVCCopyPlan) { p.Items[0].SourceName = "../data" },
		"source uid missing":        func(p *PVCCopyPlan) { p.Items[0].SourceUID = "" },
		"source uid forged":         func(p *PVCCopyPlan) { p.Items[0].SourceUID = "uid-2" },
		"source identity forged":    func(p *PVCCopyPlan) { p.Items[0].SourceIdentity = "sha256:" + strings.Repeat("c", 64) },
		"source in target":          func(p *PVCCopyPlan) { p.Items[0].SourceNamespace = p.TargetNamespace },
		"source not allowlisted":    func(p *PVCCopyPlan) { p.Items[0].SourceNamespace = "other" },
		"cross tenant":              func(p *PVCCopyPlan) { p.Items[0].SourceTenantID = "tenant-b" },
		"stale source size":         func(p *PVCCopyPlan) { p.Items[0].SourceRequestedBytes = 512 },
		"zero source size":          func(p *PVCCopyPlan) { p.Items[0].SourceRequestedBytes = 0 },
		"negative source size":      func(p *PVCCopyPlan) { p.Items[0].SourceRequestedBytes = -1 },
		"negative target size":      func(p *PVCCopyPlan) { p.Items[0].RequestedBytes = -1 },
		"target too small":          func(p *PVCCopyPlan) { p.Items[0].RequestedBytes = 512 },
		"copy bound too small":      func(p *PVCCopyPlan) { p.Items[0].MaxBytes = 512 },
		"copy bound above capacity": func(p *PVCCopyPlan) { p.Items[0].MaxBytes = 2049 },
		"negative limit":            func(p *PVCCopyPlan) { p.MaxBytes = -1 },
		"copy quota":                func(p *PVCCopyPlan) { p.MaxBytes = 1024 },
		"storage quota":             func(p *PVCCopyPlan) { p.StorageQuotaBytes = 1024 },
		"zero timeout":              func(p *PVCCopyPlan) { p.Items[0].TimeoutSeconds = 0 },
		"timeout escape":            func(p *PVCCopyPlan) { p.Items[0].TimeoutSeconds = 301 },
		"unbounded timeout":         func(p *PVCCopyPlan) { p.TimeoutSeconds = PVCCopyMaxTimeoutSeconds + 1 },
		"online method":             func(p *PVCCopyPlan) { p.Items[0].Method = "filesystem_online" },
		"invalid storage class":     func(p *PVCCopyPlan) { p.Items[0].StorageClass = "standard;id" },
		"changed storage class":     func(p *PVCCopyPlan) { p.Items[0].StorageClass = "other" },
		"unknown mode":              func(p *PVCCopyPlan) { p.Items[0].AccessModes = []string{"exec"} },
		"readonly target":           func(p *PVCCopyPlan) { p.Items[0].AccessModes = []string{"ReadOnlyMany"} },
		"duplicate mode":            func(p *PVCCopyPlan) { p.Items[0].AccessModes = []string{"ReadWriteOnce", "ReadWriteOnce"} },
		"empty selection":           func(p *PVCCopyPlan) { p.Items = nil },
		"duplicate namespace":       func(p *PVCCopyPlan) { p.AllowedSourceNamespaces = []string{"base", "base"} },
		"duplicate item":            func(p *PVCCopyPlan) { p.Items = append(p.Items, p.Items[0]) },
		"duplicate target": func(p *PVCCopyPlan) {
			item := p.Items[0]
			item.ID = "second"
			item.SourceName = "second"
			p.Items = append(p.Items, item)
		},
		"duplicate source": func(p *PVCCopyPlan) {
			item := p.Items[0]
			item.ID = "second"
			item.TargetName = "second"
			p.Items = append(p.Items, item)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			p, sources, permission := pvcCopyFixture(t)
			mutate(&p)
			if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
				t.Fatal("invalid selection accepted")
			}
		})
	}
}

func TestPVCCopyExactTrustedBinding(t *testing.T) {
	mutations := map[string]func(*PVCCopyPermission){
		"plan":        func(p *PVCCopyPermission) { p.PlanID = "other" },
		"tenant":      func(p *PVCCopyPermission) { p.TenantID = "other" },
		"project":     func(p *PVCCopyPermission) { p.ProjectID = "other" },
		"environment": func(p *PVCCopyPermission) { p.EnvironmentID = "other" },
		"revision":    func(p *PVCCopyPermission) { p.TemplateRevisionID = "other" },
		"template":    func(p *PVCCopyPermission) { p.TemplateDigest = "sha256:" + strings.Repeat("d", 64) },
		"namespace":   func(p *PVCCopyPermission) { p.TargetNamespace = "other" },
		"allowlist":   func(p *PVCCopyPermission) { p.AllowedSourceNamespaces = []string{"other"} },
		"image":       func(p *PVCCopyPermission) { p.AllowedImages = nil },
		"max bytes":   func(p *PVCCopyPermission) { p.MaxBytes = 1 },
		"timeout":     func(p *PVCCopyPermission) { p.TimeoutSeconds = 1 },
		"quota":       func(p *PVCCopyPermission) { p.StorageQuotaBytes = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			p, sources, permission := pvcCopyFixture(t)
			mutate(&permission)
			if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
				t.Fatal("mismatched trusted permission accepted")
			}
		})
	}
}

func TestPVCCopySourceScanAndApproval(t *testing.T) {
	for _, name := range []string{"missing", "duplicate", "missing UID", "block", "unclassified", "negative", "production", "cross tenant"} {
		t.Run(name, func(t *testing.T) {
			p, sources, permission := pvcCopyFixture(t)
			switch name {
			case "missing":
				sources = nil
			case "duplicate":
				sources = append(sources, sources[0])
			case "missing UID":
				sources[0].UID = ""
			case "block":
				sources[0].VolumeMode = "Block"
			case "unclassified":
				sources[0].EnvironmentClass = ""
			case "negative":
				sources[0].RequestedBytes = -1
			case "production":
				sources[0].EnvironmentClass = "production"
			case "cross tenant":
				sources[0].TenantID = "tenant-b"
			}
			if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
				t.Fatal("unsafe scanned source accepted")
			}
		})
	}
	p, sources, permission := pvcCopyFixture(t)
	sources[0].EnvironmentClass = "production"
	permission.ApprovedProductionSources = []PVCCopySourceRef{{"base", "data", "uid-old"}}
	if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("stale UID approval accepted")
	}
	permission.ApprovedProductionSources[0].UID = "uid-1"
	if _, err := CompilePVCCopyPlan(p, sources, permission); err != nil {
		t.Fatal(err)
	}
	sources[0].TenantID, p.Items[0].SourceTenantID = "tenant-b", "tenant-b"
	if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("production approval allowed cross-tenant copy")
	}
}

func TestPVCCopyDigestCannotSubstituteAuthorization(t *testing.T) {
	p, sources, permission := pvcCopyFixture(t)
	p, err := CompilePVCCopyPlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	p.Items[0].SourceUID = "forged"
	if err := p.Validate(); err == nil {
		t.Fatal("digest mutation accepted")
	}
	p.Digest, _ = p.CanonicalDigest()
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("resealed forged source accepted")
	}
	// Resizing the same UID also requires a fresh source identity and plan.
	p, sources, permission = pvcCopyFixture(t)
	p, _ = CompilePVCCopyPlan(p, sources, permission)
	sources[0].RequestedBytes++
	if err := ValidatePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("stale source scan accepted")
	}
}

func TestPVCCopyQuotaOverflow(t *testing.T) {
	p, sources, permission := pvcCopyFixture(t)
	p.MaxBytes, p.StorageQuotaBytes = math.MaxInt64, math.MaxInt64
	permission.MaxBytes, permission.StorageQuotaBytes = math.MaxInt64, math.MaxInt64
	p.Items[0].RequestedBytes, p.Items[0].MaxBytes = math.MaxInt64, math.MaxInt64
	if _, err := CompilePVCCopyPlan(p, sources, permission); err != nil {
		t.Fatal(err)
	}
	source := sources[0]
	source.Name, source.UID = "second", "uid-2"
	sources = append(sources, source)
	item := p.Items[0]
	item.ID, item.TargetName, item.SourceName, item.SourceUID = "second", "second", "second", "uid-2"
	item.RequestedBytes, item.MaxBytes = 1024, 1024
	p.Items = append(p.Items, item)
	if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("overflowed quota accepted")
	}
}

func TestPVCCopyCanonicalSetsAndOpaqueIDs(t *testing.T) {
	p, sources, permission := pvcCopyFixture(t)
	p.PlanID, permission.PlanID = "sha256:copy/preview-a", "sha256:copy/preview-a"
	p.TemplateRevisionID, permission.TemplateRevisionID = "templates/app/rev-1", "templates/app/rev-1"
	p.Items[0].ID = "PersistentVolumeClaim/base/data"
	p.AllowedSourceNamespaces = []string{"unused", "base"}
	permission.AllowedSourceNamespaces = []string{"base", "unused"}
	sources[0].AccessModes = []string{"ReadWriteOnce", "ReadWriteMany"}
	p.Items[0].AccessModes = []string{"ReadWriteMany", "ReadWriteOnce"}
	compiled, err := CompilePVCCopyPlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	digest := compiled.Digest
	compiled.AllowedSourceNamespaces[0], compiled.AllowedSourceNamespaces[1] = compiled.AllowedSourceNamespaces[1], compiled.AllowedSourceNamespaces[0]
	compiled.Items[0].AccessModes[0], compiled.Items[0].AccessModes[1] = compiled.Items[0].AccessModes[1], compiled.Items[0].AccessModes[0]
	if actual, _ := compiled.CanonicalDigest(); actual != digest {
		t.Fatal("set ordering changed plan digest")
	}
	if err := ValidatePVCCopyPlan(compiled, sources, permission); err != nil {
		t.Fatal(err)
	}
	sources[0].AccessModes[0], sources[0].AccessModes[1] = sources[0].AccessModes[1], sources[0].AccessModes[0]
	if err := ValidatePVCCopyPlan(compiled, sources, permission); err != nil {
		t.Fatal("source access mode order changed identity", err)
	}
	for _, bad := range []string{"../rev", "a/../rev", "a//rev", "a/rev/", strings.Repeat("a", 513), "rev\n"} {
		if pvcCopyID(bad) {
			t.Fatalf("unsafe opaque ID accepted: %q", bad)
		}
	}
}

func TestPVCCopyMultipleItemsQuotaAndSelection(t *testing.T) {
	p, sources, permission := pvcCopyFixture(t)
	second := sources[0]
	second.Name, second.UID = "second", "uid-2"
	sources = append(sources, second)
	item := p.Items[0]
	item.ID, item.SourceName, item.SourceUID, item.TargetName = "second", "second", "uid-2", "second-copy"
	p.Items = append(p.Items, item)
	p.MaxBytes, permission.MaxBytes = 4096, 4096
	compiled, err := CompilePVCCopyPlan(p, sources, permission)
	if err != nil {
		t.Fatal("exact quota boundary rejected", err)
	}
	compiled.Items[0], compiled.Items[1] = compiled.Items[1], compiled.Items[0]
	if err := compiled.Validate(); err != nil {
		t.Fatal("item order changed canonical digest", err)
	}
	p.StorageQuotaBytes--
	if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("aggregate target quota exceeded")
	}
	p.StorageQuotaBytes++
	p.MaxBytes--
	if _, err := CompilePVCCopyPlan(p, sources, permission); err == nil {
		t.Fatal("aggregate copy quota exceeded")
	}
}

func FuzzPVCCopyPlanJSON(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"items":[{"sourceUid":"uid-1","sourceUid":"uid-2"}]}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var plan PVCCopyPlan
		if err := json.Unmarshal(data, &plan); err == nil {
			_ = plan.Validate()
			_, _ = plan.CanonicalDigest()
		}
	})
}

func TestPVCCopyStrictJSONAndAdditiveTransport(t *testing.T) {
	p, sources, permission := pvcCopyFixture(t)
	p, err := CompilePVCCopyPlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(p)
	for _, raw := range []string{
		`null`,
		strings.Replace(string(encoded), `"planId":`, `"PlanID":`, 1),
		strings.Replace(string(encoded), `"sourceUid":`, `"sourceUID":`, 1),
		strings.Replace(string(encoded), `"contractVersion":`, `"planId":"forged","contractVersion":`, 1),
		strings.Replace(string(encoded), `"sourceUid":`, `"sourceUid":"forged","sourceUid":`, 1),
		strings.Replace(string(encoded), `"contractVersion":`, `"approved":true,"contractVersion":`, 1),
		strings.Replace(string(encoded), `"sourceUid":`, `"credentials":"forbidden","sourceUid":`, 1),
		string(encoded) + `{}`,
	} {
		var decoded PVCCopyPlan
		if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
			t.Fatal("unknown/trailing field accepted")
		}
	}
	var cmd RunnerCommand
	if err := json.Unmarshal([]byte(`{"operation":"materialize_pvc","pvcCopyPlan":`+string(encoded)+`}`), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.PVCCopyPlan == nil || cmd.PVCCopyPlan.Validate() != nil {
		t.Fatal("plan transport failed")
	}
	bad := strings.Replace(string(encoded), `"contractVersion":`, `"rows":[],"contractVersion":`, 1)
	if err := json.Unmarshal([]byte(`{"pvcCopyPlan":`+bad+`}`), &cmd); err == nil {
		t.Fatal("nested unknown fields accepted")
	}
	var legacy RunnerCommand
	if err := json.Unmarshal([]byte(`{"operation":"install"}`), &legacy); err != nil || legacy.PVCCopyPlan != nil {
		t.Fatal("legacy command broken")
	}
	for _, transport := range []struct {
		value any
		field string
	}{
		{ResourceSnapshot{SourceUID: "uid-1"}, "sourceUid"},
		{RunnerHeartbeatRequest{PVCCopyContractVersion: PVCCopyContractVersion}, "pvcCopyContractVersion"},
		{RunnerCommandResult{PVCCopyPlanDigest: p.Digest, PVCCopyVerified: true}, "pvcCopyPlanDigest"},
	} {
		data, _ := json.Marshal(transport.value)
		if !strings.Contains(string(data), `"`+transport.field+`"`) {
			t.Fatal("missing additive metadata field")
		}
	}
}
