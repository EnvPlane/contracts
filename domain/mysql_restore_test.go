package domain

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

func mysqlRestoreFixture(t *testing.T) (MySQLRestorePlan, []MySQLRestoreSource, MySQLRestorePermission) {
	t.Helper()
	digest := "sha256:" + strings.Repeat("a", 64)
	s := MySQLRestoreSource{TenantID: "tenant", Namespace: "base", PVCName: "mysql-data", PVCUID: "pvc-uid", WorkloadKind: "StatefulSet", WorkloadName: "mysql", WorkloadUID: "workload-uid", Container: "mysql", Database: "app", Username: "reader", SecretName: "mysql-auth", SecretUID: "secret-uid", PasswordKey: "password", SourceImage: "registry/mysql@" + digest, StorageClass: "standard", AccessModes: []string{"ReadWriteOnce"}, RequestedBytes: 1024, EnvironmentClass: "development"} // #nosec G101 -- password is a Kubernetes Secret key name; no credential value is present.
	s.Service, s.ServiceUID, s.Port = "mysql", "service-uid", 3306
	s.BackupAdminSecretRef = &MySQLRestoreCredentialRef{Namespace: "base", Name: "mysql-backup", UID: "backup-uid", Username: "backup", PasswordKey: "password"}
	s.TLS = MySQLRestoreTLS{CASecretName: "mysql-ca", CASecretUID: "ca-uid", CAKey: "ca.crt", ServerName: "mysql.base.svc.cluster.local"}
	item := MySQLRestoreItem{ID: "mysql", Source: s, TargetPVCName: "mysql-new", TargetSecretName: "mysql-generated", TargetSecretUID: "generated-uid", TargetPasswordKey: "password", TargetDatabase: "app", TargetUsername: "target", SecretMaterializationPlanDigest: digest, StorageClass: "standard", AccessModes: []string{"ReadWriteOnce"}, RequestedBytes: 2048, MaxBytes: 2048, TimeoutSeconds: 300, DDLMode: "backup_lock", MaxStatementBytes: 1024, MaxTables: 100, MaxRows: 10000}
	item.TargetRootPasswordKey = MySQLRestoreRootPasswordKey
	p := MySQLRestorePlan{PlanID: "restore-1", TenantID: "tenant", ProjectID: "project", EnvironmentID: "environment", EnvironmentCreatedAt: time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC), TemplateRevisionID: "revision", TemplateDigest: digest, TargetNamespace: "preview", AllowedSourceNamespaces: []string{"base"}, HelperImage: "registry/helper@" + digest, TargetImage: "registry/mysql@" + digest, Items: []MySQLRestoreItem{item}, MaxBytes: 2048, TimeoutSeconds: 300, StorageQuotaBytes: 2048}
	permission := MySQLRestorePermission{PlanID: p.PlanID, TenantID: p.TenantID, ProjectID: p.ProjectID, EnvironmentID: p.EnvironmentID, EnvironmentCreatedAt: p.EnvironmentCreatedAt, TemplateRevisionID: p.TemplateRevisionID, TemplateDigest: p.TemplateDigest, TargetNamespace: p.TargetNamespace, AllowedSourceNamespaces: []string{"base"}, AllowedHelperImages: []string{p.HelperImage}, AllowedTargetImages: []string{p.TargetImage}, MaxBytes: p.MaxBytes, TimeoutSeconds: p.TimeoutSeconds, StorageQuotaBytes: p.StorageQuotaBytes, TargetSecrets: []MySQLRestoreTargetSecret{{Name: item.TargetSecretName, UID: item.TargetSecretUID, PasswordKey: item.TargetPasswordKey, Username: item.TargetUsername, SecretMaterializationPlanDigest: item.SecretMaterializationPlanDigest, RootPasswordKey: item.TargetRootPasswordKey}}}
	permission.ApprovedBackupAdminSecrets = []MySQLRestoreCredentialRef{*s.BackupAdminSecretRef}
	return p, []MySQLRestoreSource{s}, permission
}

func TestMySQLRestoreCompileTransportAndLifecycle(t *testing.T) {
	p, sources, permission := mysqlRestoreFixture(t)
	compiled, err := CompileMySQLRestorePlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	if p.Items[0].SourceIdentity != "" {
		t.Fatal("compiler mutated caller")
	}
	encoded, _ := json.Marshal(RunnerCommand{Operation: RunnerOperationRestoreMySQL, MySQLRestorePlan: &compiled})
	var decoded RunnerCommand
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMySQLRestorePlan(*decoded.MySQLRestorePlan, sources, permission); err != nil {
		t.Fatal(err)
	}
	r := MySQLRestoreResult{ContractVersion: MySQLRestoreContractVersion, CommandID: "command", AttemptID: "attempt", PlanID: compiled.PlanID, PlanDigest: compiled.Digest, TenantID: compiled.TenantID, ProjectID: compiled.ProjectID, EnvironmentID: compiled.EnvironmentID, EnvironmentCreatedAt: compiled.EnvironmentCreatedAt, Verified: true}
	if err := r.Validate(compiled, "command", "attempt"); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*MySQLRestoreResult){func(r *MySQLRestoreResult) { r.Verified = false }, func(r *MySQLRestoreResult) { r.CommandID = "old" }, func(r *MySQLRestoreResult) { r.AttemptID = "old" }, func(r *MySQLRestoreResult) { r.PlanDigest = "sha256:" + strings.Repeat("b", 64) }, func(r *MySQLRestoreResult) { r.EnvironmentCreatedAt = r.EnvironmentCreatedAt.Add(-time.Hour) }} {
		bad := r
		mutate(&bad)
		if err := bad.Validate(compiled, "command", "attempt"); err == nil {
			t.Fatal("stale evidence accepted")
		}
	}
	newInstance := compiled
	newInstance.EnvironmentCreatedAt = newInstance.EnvironmentCreatedAt.Add(time.Second)
	newInstance.Digest, _ = newInstance.CanonicalDigest()
	if newInstance.Digest == compiled.Digest {
		t.Fatal("lifecycle not covered by digest")
	}
	if err := r.Validate(newInstance, "command", "attempt"); err == nil {
		t.Fatal("old lifecycle unblocked reused environment ID")
	}
	permission.EnvironmentCreatedAt = permission.EnvironmentCreatedAt.Add(time.Second)
	if err := ValidateMySQLRestorePlan(compiled, sources, permission); err == nil {
		t.Fatal("lifecycle policy mismatch accepted")
	}
}

func TestMySQLRestoreRejectsForgedAndUnsafeMetadata(t *testing.T) {
	tests := map[string]func(*MySQLRestorePlan){
		"missing lifecycle":       func(p *MySQLRestorePlan) { p.EnvironmentCreatedAt = time.Time{} },
		"missing pvc UID":         func(p *MySQLRestorePlan) { p.Items[0].Source.PVCUID = "" },
		"forged pvc UID":          func(p *MySQLRestorePlan) { p.Items[0].Source.PVCUID = "forged" },
		"forged workload UID":     func(p *MySQLRestorePlan) { p.Items[0].Source.WorkloadUID = "forged" },
		"forged secret UID":       func(p *MySQLRestorePlan) { p.Items[0].Source.SecretUID = "forged" },
		"unknown class":           func(p *MySQLRestorePlan) { p.Items[0].Source.EnvironmentClass = "" },
		"cross tenant":            func(p *MySQLRestorePlan) { p.Items[0].Source.TenantID = "other" },
		"production flag":         func(p *MySQLRestorePlan) { p.Items[0].Source.EnvironmentClass = "production" },
		"source namespace":        func(p *MySQLRestorePlan) { p.Items[0].Source.Namespace = p.TargetNamespace },
		"database SQL":            func(p *MySQLRestorePlan) { p.Items[0].Source.Database = "app;DROP DATABASE app" },
		"database option":         func(p *MySQLRestorePlan) { p.Items[0].TargetDatabase = "--all-databases" },
		"shell target":            func(p *MySQLRestorePlan) { p.Items[0].TargetPVCName = "data$(id)" },
		"source image":            func(p *MySQLRestorePlan) { p.Items[0].Source.SourceImage = "mysql:latest" },
		"target image":            func(p *MySQLRestorePlan) { p.TargetImage = "mysql:latest" },
		"helper allowlist":        func(p *MySQLRestorePlan) { p.HelperImage = "registry/other@sha256:" + strings.Repeat("a", 64) },
		"copied secret":           func(p *MySQLRestorePlan) { p.Items[0].TargetSecretUID = p.Items[0].Source.SecretUID },
		"forged generated secret": func(p *MySQLRestorePlan) { p.Items[0].TargetSecretUID = "forged" },
		"forged secret plan": func(p *MySQLRestorePlan) {
			p.Items[0].SecretMaterializationPlanDigest = "sha256:" + strings.Repeat("b", 64)
		},
		"stale size":           func(p *MySQLRestorePlan) { p.Items[0].Source.RequestedBytes = 512 },
		"negative target size": func(p *MySQLRestorePlan) { p.Items[0].RequestedBytes = -1 },
		"negative bytes":       func(p *MySQLRestorePlan) { p.Items[0].MaxBytes = -1 },
		"quota overflow":       func(p *MySQLRestorePlan) { p.Items[0].RequestedBytes = math.MaxInt64 },
		"timeout":              func(p *MySQLRestorePlan) { p.Items[0].TimeoutSeconds = 301 },
		"duplicate selection":  func(p *MySQLRestorePlan) { p.Items = append(p.Items, p.Items[0]) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			p, sources, permission := mysqlRestoreFixture(t)
			mutate(&p)
			if _, err := CompileMySQLRestorePlan(p, sources, permission); err == nil {
				t.Fatal("unsafe source/plan accepted")
			}
		})
	}
}

func TestMySQLRestoreIndependentRootPasswordBinding(t *testing.T) {
	for _, key := range []string{"", "password", "../root", "root;id", "root\n", strings.Repeat("r", 254)} {
		t.Run(key, func(t *testing.T) {
			p, sources, permission := mysqlRestoreFixture(t)
			p.Items[0].TargetRootPasswordKey = key
			if _, err := CompileMySQLRestorePlan(p, sources, permission); err == nil {
				t.Fatal("missing, reused or invalid root password key accepted")
			}
		})
	}
	p, sources, permission := mysqlRestoreFixture(t)
	compiled, err := CompileMySQLRestorePlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(compiled)
	var decoded MySQLRestorePlan
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Items[0].TargetRootPasswordKey != MySQLRestoreRootPasswordKey {
		t.Fatal("transport dropped independent root key")
	}
	if err := ValidateMySQLRestorePlan(decoded, sources, permission); err != nil {
		t.Fatal(err)
	}
	changed := compiled
	changed.Items = append([]MySQLRestoreItem(nil), compiled.Items...)
	changed.Items[0].TargetRootPasswordKey = "OTHER_ROOT_KEY"
	if err := changed.Validate(); err == nil {
		t.Fatal("root key mutation retained immutable digest")
	}
	changed.Digest, _ = changed.CanonicalDigest()
	if changed.Digest == compiled.Digest {
		t.Fatal("root key omitted from canonical digest")
	}
	if err := changed.Validate(); err != nil {
		t.Fatal("distinct syntactically valid root key should require out-of-band policy check", err)
	}
	if err := ValidateMySQLRestorePlan(changed, sources, permission); err == nil {
		t.Fatal("re-sealed unapproved root key accepted")
	}
	permission.TargetSecrets[0].RootPasswordKey = ""
	if err := ValidateMySQLRestorePlan(compiled, sources, permission); err == nil {
		t.Fatal("trusted Secret gate missing root key accepted")
	}
	permission.TargetSecrets[0].RootPasswordKey = "OTHER_ROOT_KEY"
	if err := ValidateMySQLRestorePlan(compiled, sources, permission); err == nil {
		t.Fatal("mismatched trusted Secret root key accepted")
	}
	// The unpublished MySQL schema deliberately refuses old plans lacking the
	// root binding rather than silently defaulting back to the app password.
	withoutRoot := strings.Replace(string(encoded), `"targetRootPasswordKey":"MYSQL_ROOT_PASSWORD",`, "", 1)
	if err := json.Unmarshal([]byte(withoutRoot), &decoded); err != nil {
		t.Fatal(err)
	}
	decoded.Digest, _ = decoded.CanonicalDigest()
	if err := decoded.Validate(); err == nil {
		t.Fatal("old MySQL plan without root binding accepted")
	}
}

func TestMySQLRestoreReviewedDDLAndTLS(t *testing.T) {
	for _, name := range []string{"Deployment", "service UID", "port", "admin missing", "admin approval", "admin alias", "CA missing", "wrong DNS", "statement", "tables", "rows", "root", "system database", "image mismatch"} {
		t.Run(name, func(t *testing.T) {
			p, sources, permission := mysqlRestoreFixture(t)
			switch name {
			case "Deployment":
				p.Items[0].Source.WorkloadKind = "Deployment"
			case "service UID":
				p.Items[0].Source.ServiceUID = ""
			case "port":
				p.Items[0].Source.Port = 65536
			case "admin missing":
				p.Items[0].Source.BackupAdminSecretRef = nil
			case "admin approval":
				permission.ApprovedBackupAdminSecrets = nil
			case "admin alias":
				r := *p.Items[0].Source.BackupAdminSecretRef
				r.UID = p.Items[0].Source.SecretUID
				p.Items[0].Source.BackupAdminSecretRef = &r
			case "CA missing":
				p.Items[0].Source.TLS.CASecretUID = ""
			case "wrong DNS":
				p.Items[0].Source.TLS.ServerName = "mysql.base.svc-evil.example"
			case "statement":
				p.Items[0].MaxStatementBytes = (16 << 20) + 1
			case "tables":
				p.Items[0].MaxTables = 1001
			case "rows":
				p.Items[0].MaxRows = 100000001
			case "root":
				p.Items[0].Source.Username = "root"
			case "system database":
				p.Items[0].Source.Database = "MYSQL"
			case "image mismatch":
				p.TargetImage = "registry/mysql@sha256:" + strings.Repeat("b", 64)
				permission.AllowedTargetImages = []string{p.TargetImage}
			}
			if _, err := CompileMySQLRestorePlan(p, sources, permission); err == nil {
				t.Fatal("unreviewed/unsafe MySQL mode accepted")
			}
		})
	}
	p, sources, permission := mysqlRestoreFixture(t)
	p.Items[0].DDLMode = "quiescence"
	p.Items[0].QuiescenceRef = "authority/lease-1"
	p.Items[0].Source.BackupAdminSecretRef = nil
	sources[0].BackupAdminSecretRef = nil
	if _, err := CompileMySQLRestorePlan(p, sources, permission); err == nil {
		t.Fatal("unapproved quiescence accepted")
	}
	permission.ApprovedQuiescenceRefs = []string{"authority/lease-1"}
	compiled, err := CompileMySQLRestorePlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(compiled)
	var decoded MySQLRestorePlan
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMySQLRestorePlan(decoded, sources, permission); err != nil {
		t.Fatal(err)
	}
	p, sources, permission = mysqlRestoreFixture(t)
	compiled, err = CompileMySQLRestorePlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	compiled.Items[0].Source.BackupAdminSecretRef.UID = "changed"
	if p.Items[0].Source.BackupAdminSecretRef.UID != "backup-uid" {
		t.Fatal("compiler aliases backup credential refs")
	}
}

func TestMySQLRestoreProductionAndStrictJSON(t *testing.T) {
	p, sources, permission := mysqlRestoreFixture(t)
	sources[0].EnvironmentClass = "production"
	p.Items[0].Source.EnvironmentClass = "production"
	if _, err := CompileMySQLRestorePlan(p, sources, permission); err == nil {
		t.Fatal("unapproved production accepted")
	}
	permission.ApprovedProductionSources = []PVCCopySourceRef{{sources[0].Namespace, sources[0].PVCName, sources[0].PVCUID}}
	compiled, err := CompileMySQLRestorePlan(p, sources, permission)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(compiled)
	for _, bad := range []string{`null`, strings.Replace(string(encoded), `"planId":`, `"credentials":"no","planId":`, 1), strings.Replace(string(encoded), `"pvcUid":`, `"password":"no","pvcUid":`, 1), strings.Replace(string(encoded), `"secretUid":`, `"secretUid":"forged","secretUid":`, 1), strings.Replace(string(encoded), `"database":`, `"Database":`, 1)} {
		var decoded RunnerCommand
		if err := json.Unmarshal([]byte(`{"mysqlRestorePlan":`+bad+`}`), &decoded); err == nil && bad != `null` {
			t.Fatal("unknown, duplicate or case alias metadata accepted")
		}
	}
	sources[0].SecretUID = "replacement"
	if err := ValidateMySQLRestorePlan(compiled, sources, permission); err == nil {
		t.Fatal("replaced credential Secret accepted")
	}
}
