package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

const MySQLRestoreContractVersion = "v1"
const RunnerOperationRestoreMySQL = "restore_mysql"
const MySQLRestoreMaxTimeoutSeconds int64 = 21600
const MySQLRestoreMaxBytes int64 = 1 << 40

// MySQLRestoreTargetDataSubPath is shared by temporary target initialization
// and the generated feature workload; source PVC root mounts are unchanged.
const MySQLRestoreTargetDataSubPath = "mysql"

const MySQLRestoreRootPasswordKey = "MYSQL_ROOT_PASSWORD" // #nosec G101 -- fixed Kubernetes Secret key name, never a credential value.

type MySQLRestoreCredentialRef struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	UID         string `json:"uid"`
	Username    string `json:"username,omitempty"`
	UsernameKey string `json:"usernameKey,omitempty"`
	PasswordKey string `json:"passwordKey"`
}

type MySQLRestoreTLS struct {
	CASecretName string `json:"caSecretName"`
	CASecretUID  string `json:"caSecretUid"`
	CAKey        string `json:"caKey"`
	ServerName   string `json:"serverName"`
}

// MySQLRestoreSource is trusted scan metadata, never credential values or dumps.
// Secret UID is preserved separately from the sanitized Secret manifest.
type MySQLRestoreSource struct {
	Service              string                     `json:"service"`
	ServiceUID           string                     `json:"serviceUid"`
	Port                 int                        `json:"port"`
	BackupAdminSecretRef *MySQLRestoreCredentialRef `json:"backupAdminSecretRef,omitempty"`
	TLS                  MySQLRestoreTLS            `json:"tls"`
	UsernameKey          string                     `json:"usernameKey,omitempty"`
	TenantID             string                     `json:"tenantId"`
	Namespace            string                     `json:"namespace"`
	PVCName              string                     `json:"pvcName"`
	PVCUID               string                     `json:"pvcUid"`
	WorkloadKind         string                     `json:"workloadKind"`
	WorkloadName         string                     `json:"workloadName"`
	WorkloadUID          string                     `json:"workloadUid"`
	Container            string                     `json:"container"`
	Database             string                     `json:"database"`
	Username             string                     `json:"username"`
	SecretName           string                     `json:"secretName"`
	SecretUID            string                     `json:"secretUid"`
	PasswordKey          string                     `json:"passwordKey"`
	SourceImage          string                     `json:"sourceImage"`
	StorageClass         string                     `json:"storageClass"`
	AccessModes          []string                   `json:"accessModes"`
	RequestedBytes       int64                      `json:"requestedBytes"`
	EnvironmentClass     string                     `json:"environmentClass"`
}

type MySQLRestoreItem struct {
	DDLMode                         string             `json:"ddlMode"`
	QuiescenceRef                   string             `json:"quiescenceRef,omitempty"`
	MaxStatementBytes               int                `json:"maxStatementBytes"`
	MaxTables                       int                `json:"maxTables"`
	MaxRows                         int64              `json:"maxRows"`
	ID                              string             `json:"id"`
	Source                          MySQLRestoreSource `json:"source"`
	SourceIdentity                  string             `json:"sourceIdentity"`
	TargetPVCName                   string             `json:"targetPvcName"`
	TargetSecretName                string             `json:"targetSecretName"`
	TargetSecretUID                 string             `json:"targetSecretUid"`
	TargetPasswordKey               string             `json:"targetPasswordKey"`
	TargetRootPasswordKey           string             `json:"targetRootPasswordKey"`
	TargetDatabase                  string             `json:"targetDatabase"`
	TargetUsername                  string             `json:"targetUsername"`
	SecretMaterializationPlanDigest string             `json:"secretMaterializationPlanDigest"`
	StorageClass                    string             `json:"storageClass"`
	AccessModes                     []string           `json:"accessModes"`
	RequestedBytes                  int64              `json:"requestedBytes"`
	MaxBytes                        int64              `json:"maxBytes"`
	TimeoutSeconds                  int64              `json:"timeoutSeconds"`
}

// EnvironmentCreatedAt is the immutable lifecycle instance timestamp, not the
// time compilation happened. No credentials, rows, dumps or command text exist.
type MySQLRestorePlan struct {
	ContractVersion         string             `json:"contractVersion"`
	PlanID                  string             `json:"planId"`
	TenantID                string             `json:"tenantId"`
	ProjectID               string             `json:"projectId"`
	EnvironmentID           string             `json:"environmentId"`
	EnvironmentCreatedAt    time.Time          `json:"environmentCreatedAt"`
	TemplateRevisionID      string             `json:"templateRevisionId"`
	TemplateDigest          string             `json:"templateDigest"`
	TargetNamespace         string             `json:"targetNamespace"`
	AllowedSourceNamespaces []string           `json:"allowedSourceNamespaces"`
	HelperImage             string             `json:"helperImage"`
	TargetImage             string             `json:"targetImage"`
	Items                   []MySQLRestoreItem `json:"items"`
	MaxBytes                int64              `json:"maxBytes"`
	TimeoutSeconds          int64              `json:"timeoutSeconds"`
	StorageQuotaBytes       int64              `json:"storageQuotaBytes"`
	Digest                  string             `json:"digest"`
}

// MySQLRestorePermission is never accepted from clients as an approval flag.
// Target bindings must come from the completed generated Secret gate.
type MySQLRestorePermission struct {
	ApprovedBackupAdminSecrets                                        []MySQLRestoreCredentialRef
	ApprovedQuiescenceRefs                                            []string
	PlanID, TenantID, ProjectID, EnvironmentID                        string
	EnvironmentCreatedAt                                              time.Time
	TemplateRevisionID, TemplateDigest, TargetNamespace               string
	AllowedSourceNamespaces, AllowedHelperImages, AllowedTargetImages []string
	ApprovedProductionSources                                         []PVCCopySourceRef
	TargetSecrets                                                     []MySQLRestoreTargetSecret
	MaxBytes, TimeoutSeconds, StorageQuotaBytes                       int64
}

type MySQLRestoreTargetSecret struct {
	Name, UID, PasswordKey, Username, SecretMaterializationPlanDigest string
	RootPasswordKey                                                   string
}

// MySQLRestoreResult is prerequisite evidence, not environment Ready. Validate
// against the CURRENT leased command/attempt and current lifecycle-bound plan.
type MySQLRestoreResult struct {
	ContractVersion      string    `json:"contractVersion"`
	CommandID            string    `json:"commandId"`
	AttemptID            string    `json:"attemptId"`
	PlanID               string    `json:"planId"`
	PlanDigest           string    `json:"planDigest"`
	TenantID             string    `json:"tenantId"`
	ProjectID            string    `json:"projectId"`
	EnvironmentID        string    `json:"environmentId"`
	EnvironmentCreatedAt time.Time `json:"environmentCreatedAt"`
	Verified             bool      `json:"verified"`
}

var mysqlRestoreIdentifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)
var mysqlRestoreSecretKey = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,252}$`)

func (r MySQLRestoreCredentialRef) valid() bool {
	return pvcCopyName(r.Namespace, true) && pvcCopyName(r.Name, false) && pvcCopyUIDPattern.MatchString(r.UID) && mysqlRestoreSecretKey.MatchString(r.PasswordKey) && ((mysqlRestoreIdentifier.MatchString(r.Username) && r.UsernameKey == "") || (r.Username == "" && mysqlRestoreSecretKey.MatchString(r.UsernameKey)))
}

func mysqlRestoreDatabase(s string) bool {
	if !mysqlRestoreIdentifier.MatchString(s) {
		return false
	}
	switch strings.ToLower(s) {
	case "mysql", "sys", "information_schema", "performance_schema":
		return false
	}
	return true
}

func (s MySQLRestoreSource) CanonicalIdentity() (string, error) {
	pvc := PVCCopySource{TenantID: s.TenantID, Namespace: s.Namespace, Name: s.PVCName, UID: s.PVCUID, StorageClass: s.StorageClass, AccessModes: s.AccessModes, RequestedBytes: s.RequestedBytes, VolumeMode: "Filesystem", EnvironmentClass: s.EnvironmentClass}
	if _, err := pvc.CanonicalIdentity(); err != nil {
		return "", err
	}
	if s.WorkloadKind != "StatefulSet" || !pvcCopyName(s.WorkloadName, false) || !pvcCopyUIDPattern.MatchString(s.WorkloadUID) || !pvcCopyName(s.Container, true) || !mysqlRestoreDatabase(s.Database) || !pvcCopyImagePattern.MatchString(s.SourceImage) || !pvcCopyName(s.Service, true) || !pvcCopyUIDPattern.MatchString(s.ServiceUID) || s.Port < 1 || s.Port > 65535 {
		return "", errors.New("invalid immutable MySQL source metadata")
	}
	app := MySQLRestoreCredentialRef{s.Namespace, s.SecretName, s.SecretUID, s.Username, s.UsernameKey, s.PasswordKey}
	if !app.valid() || strings.EqualFold(s.Username, "root") {
		return "", errors.New("invalid MySQL application credential metadata")
	}
	if s.BackupAdminSecretRef != nil && (!s.BackupAdminSecretRef.valid() || s.BackupAdminSecretRef.Namespace != s.Namespace || s.BackupAdminSecretRef.UID == s.SecretUID || s.BackupAdminSecretRef.Name == s.SecretName) {
		return "", errors.New("invalid distinct backup-admin credential metadata")
	}
	dns := s.Service + "." + s.Namespace + ".svc"
	if !pvcCopyName(s.TLS.CASecretName, false) || !pvcCopyUIDPattern.MatchString(s.TLS.CASecretUID) || !mysqlRestoreSecretKey.MatchString(s.TLS.CAKey) || !pvcCopyName(s.TLS.ServerName, false) || (s.TLS.ServerName != dns && s.TLS.ServerName != dns+".cluster.local") {
		return "", errors.New("MySQL source requires reviewed CA Secret and Service DNS TLS identity")
	}
	s.AccessModes = append([]string(nil), s.AccessModes...)
	sort.Strings(s.AccessModes)
	return canonicalDigest(&s, func(any) {})
}

func (p MySQLRestorePlan) CanonicalDigest() (string, error) {
	p.EnvironmentCreatedAt = p.EnvironmentCreatedAt.UTC()
	p.AllowedSourceNamespaces = append([]string(nil), p.AllowedSourceNamespaces...)
	sort.Strings(p.AllowedSourceNamespaces)
	p.Items = append([]MySQLRestoreItem(nil), p.Items...)
	for i := range p.Items {
		p.Items[i].AccessModes = append([]string(nil), p.Items[i].AccessModes...)
		sort.Strings(p.Items[i].AccessModes)
		p.Items[i].Source.AccessModes = append([]string(nil), p.Items[i].Source.AccessModes...)
		sort.Strings(p.Items[i].Source.AccessModes)
	}
	sort.Slice(p.Items, func(i, j int) bool { return p.Items[i].ID < p.Items[j].ID })
	return canonicalDigest(&p, func(v any) { v.(*MySQLRestorePlan).Digest = "" })
}

func (p MySQLRestorePlan) Validate() error {
	if p.ContractVersion != MySQLRestoreContractVersion || p.EnvironmentCreatedAt.IsZero() || !pvcCopyDigestPattern.MatchString(p.TemplateDigest) || !pvcCopyDigestPattern.MatchString(p.Digest) || !pvcCopyName(p.TargetNamespace, true) || !pvcCopyImagePattern.MatchString(p.HelperImage) || !pvcCopyImagePattern.MatchString(p.TargetImage) {
		return errors.New("invalid MySQL restore plan/lifecycle binding")
	}
	for _, id := range []string{p.PlanID, p.TenantID, p.ProjectID, p.EnvironmentID, p.TemplateRevisionID} {
		if !pvcCopyID(id) {
			return errors.New("invalid MySQL restore identity")
		}
	}
	if p.MaxBytes <= 0 || p.StorageQuotaBytes <= 0 || p.TimeoutSeconds <= 0 || p.TimeoutSeconds > MySQLRestoreMaxTimeoutSeconds || len(p.Items) == 0 {
		return errors.New("MySQL restore requires bounded limits and items")
	}
	namespaces := map[string]bool{}
	for _, ns := range p.AllowedSourceNamespaces {
		if !pvcCopyName(ns, true) || ns == p.TargetNamespace || namespaces[ns] {
			return errors.New("invalid MySQL source namespace allowlist")
		}
		namespaces[ns] = true
	}
	ids, targets, secrets, sources := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	var storage, copied int64
	for _, item := range p.Items {
		s := item.Source
		key := s.Namespace + "/" + s.PVCName
		identity, err := s.CanonicalIdentity()
		if err != nil {
			return err
		}
		if !pvcCopyID(item.ID) || ids[item.ID] || targets[item.TargetPVCName] || secrets[item.TargetSecretName] || sources[key] || s.TenantID != p.TenantID || !namespaces[s.Namespace] || identity != item.SourceIdentity {
			return errors.New("ambiguous or mismatched MySQL source selection")
		}
		ids[item.ID], targets[item.TargetPVCName], secrets[item.TargetSecretName], sources[key] = true, true, true, true
		if !mysqlRestoreSecretKey.MatchString(item.TargetRootPasswordKey) || item.TargetRootPasswordKey == item.TargetPasswordKey {
			return errors.New("MySQL target root password requires a distinct valid Secret key")
		}
		if !pvcCopyName(item.TargetPVCName, false) || !pvcCopyName(item.TargetSecretName, false) || !pvcCopyUIDPattern.MatchString(item.TargetSecretUID) || item.TargetSecretUID == s.SecretUID || !mysqlRestoreSecretKey.MatchString(item.TargetPasswordKey) || !mysqlRestoreDatabase(item.TargetDatabase) || item.TargetDatabase != s.Database || p.TargetImage != s.SourceImage || !mysqlRestoreIdentifier.MatchString(item.TargetUsername) || strings.EqualFold(item.TargetUsername, "root") || !pvcCopyDigestPattern.MatchString(item.SecretMaterializationPlanDigest) || !pvcCopyName(item.StorageClass, false) || !pvcCopyModes(item.AccessModes) || pvcCopyContains(item.AccessModes, "ReadOnlyMany") {
			return errors.New("invalid MySQL target or generated Secret binding")
		}
		if item.RequestedBytes < s.RequestedBytes || item.RequestedBytes > MySQLRestoreMaxBytes || item.MaxBytes <= 0 || item.MaxBytes > MySQLRestoreMaxBytes || item.MaxBytes > item.RequestedBytes || item.TimeoutSeconds <= 0 || item.TimeoutSeconds > p.TimeoutSeconds || item.RequestedBytes > p.StorageQuotaBytes-storage || item.MaxBytes > p.MaxBytes-copied || item.MaxStatementBytes < 1024 || item.MaxStatementBytes > 16<<20 || item.MaxTables < 1 || item.MaxTables > 1000 || item.MaxRows < 1 || item.MaxRows > 100000000 {
			return errors.New("MySQL restore exceeds size/time/quota bounds")
		}
		switch item.DDLMode {
		case "backup_lock":
			if s.BackupAdminSecretRef == nil || !s.BackupAdminSecretRef.valid() || item.QuiescenceRef != "" {
				return errors.New("backup_lock requires reviewed separate admin Secret")
			}
		case "quiescence":
			if !pvcCopyID(item.QuiescenceRef) || len(item.QuiescenceRef) > 256 {
				return errors.New("MySQL quiescence requires exact trusted authority reference")
			}
		default:
			return errors.New("unsupported MySQL DDL protection mode")
		}
		storage += item.RequestedBytes
		copied += item.MaxBytes
	}
	actual, err := p.CanonicalDigest()
	if err != nil {
		return err
	}
	if actual != p.Digest {
		return errors.New("MySQL restore digest mismatch")
	}
	return nil
}

func ValidateMySQLRestorePlan(p MySQLRestorePlan, sources []MySQLRestoreSource, permission MySQLRestorePermission) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.PlanID != permission.PlanID || p.TenantID != permission.TenantID || p.ProjectID != permission.ProjectID || p.EnvironmentID != permission.EnvironmentID || !p.EnvironmentCreatedAt.Equal(permission.EnvironmentCreatedAt) || p.TemplateRevisionID != permission.TemplateRevisionID || p.TemplateDigest != permission.TemplateDigest || p.TargetNamespace != permission.TargetNamespace || !pvcCopySameSet(p.AllowedSourceNamespaces, permission.AllowedSourceNamespaces) {
		return errors.New("MySQL restore permission/lifecycle mismatch")
	}
	if !pvcCopyContains(permission.AllowedHelperImages, p.HelperImage) || !pvcCopyContains(permission.AllowedTargetImages, p.TargetImage) || permission.MaxBytes <= 0 || p.MaxBytes > permission.MaxBytes || permission.TimeoutSeconds <= 0 || p.TimeoutSeconds > permission.TimeoutSeconds || permission.StorageQuotaBytes <= 0 || p.StorageQuotaBytes > permission.StorageQuotaBytes {
		return errors.New("MySQL image or bounds not permitted")
	}
	index := map[string]MySQLRestoreSource{}
	for _, s := range sources {
		key := s.Namespace + "/" + s.PVCName
		if _, exists := index[key]; exists {
			return errors.New("ambiguous MySQL scan source")
		}
		index[key] = s
	}
	for _, item := range p.Items {
		s, found := index[item.Source.Namespace+"/"+item.Source.PVCName]
		if !found {
			return errors.New("MySQL scanned source missing")
		}
		identity, err := s.CanonicalIdentity()
		if err != nil {
			return err
		}
		if identity != item.SourceIdentity || s.TenantID != p.TenantID {
			return errors.New("MySQL scanned immutable source mismatch")
		}
		if s.EnvironmentClass == "production" {
			approved := false
			for _, ref := range permission.ApprovedProductionSources {
				if ref == (PVCCopySourceRef{s.Namespace, s.PVCName, s.PVCUID}) {
					approved = true
				}
			}
			if !approved {
				return errors.New("production MySQL source requires exact trusted approval")
			}
		}
		if item.DDLMode == "backup_lock" {
			approved := false
			for _, ref := range permission.ApprovedBackupAdminSecrets {
				if s.BackupAdminSecretRef != nil && ref == *s.BackupAdminSecretRef {
					approved = true
				}
			}
			if !approved {
				return errors.New("backup admin Secret lacks trusted onboarding approval")
			}
		}
		if item.DDLMode == "quiescence" && !pvcCopyContains(permission.ApprovedQuiescenceRefs, item.QuiescenceRef) {
			return errors.New("MySQL quiescence authority reference not permitted")
		}
		approved := false
		for _, secret := range permission.TargetSecrets {
			if secret == (MySQLRestoreTargetSecret{Name: item.TargetSecretName, UID: item.TargetSecretUID, PasswordKey: item.TargetPasswordKey, Username: item.TargetUsername, SecretMaterializationPlanDigest: item.SecretMaterializationPlanDigest, RootPasswordKey: item.TargetRootPasswordKey}) {
				approved = true
			}
		}
		if !approved {
			return errors.New("MySQL target Secret was not materialized by the trusted gate")
		}
	}
	return nil
}

func CompileMySQLRestorePlan(p MySQLRestorePlan, sources []MySQLRestoreSource, permission MySQLRestorePermission) (MySQLRestorePlan, error) {
	if p.ContractVersion == "" {
		p.ContractVersion = MySQLRestoreContractVersion
	}
	p.EnvironmentCreatedAt = p.EnvironmentCreatedAt.UTC()
	p.AllowedSourceNamespaces = append([]string(nil), p.AllowedSourceNamespaces...)
	p.Items = append([]MySQLRestoreItem(nil), p.Items...)
	for i := range p.Items {
		if p.Items[i].Source.BackupAdminSecretRef != nil {
			ref := *p.Items[i].Source.BackupAdminSecretRef
			p.Items[i].Source.BackupAdminSecretRef = &ref
		}
		p.Items[i].AccessModes = append([]string(nil), p.Items[i].AccessModes...)
		p.Items[i].Source.AccessModes = append([]string(nil), p.Items[i].Source.AccessModes...)
		if p.Items[i].SourceIdentity == "" {
			identity, err := p.Items[i].Source.CanonicalIdentity()
			if err != nil {
				return MySQLRestorePlan{}, err
			}
			p.Items[i].SourceIdentity = identity
		}
	}
	digest, err := p.CanonicalDigest()
	if err != nil {
		return MySQLRestorePlan{}, err
	}
	if p.Digest != "" && p.Digest != digest {
		return MySQLRestorePlan{}, errors.New("MySQL restore input digest mismatch")
	}
	p.Digest = digest
	if err := ValidateMySQLRestorePlan(p, sources, permission); err != nil {
		return MySQLRestorePlan{}, err
	}
	return p, nil
}

func (r MySQLRestoreResult) Validate(p MySQLRestorePlan, commandID, attemptID string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !pvcCopyID(commandID) || !pvcCopyID(attemptID) || r.ContractVersion != MySQLRestoreContractVersion || r.CommandID != commandID || r.AttemptID != attemptID || r.PlanID != p.PlanID || r.PlanDigest != p.Digest || r.TenantID != p.TenantID || r.ProjectID != p.ProjectID || r.EnvironmentID != p.EnvironmentID || !r.EnvironmentCreatedAt.Equal(p.EnvironmentCreatedAt) || !r.Verified {
		return errors.New("unverified or stale MySQL restore result")
	}
	return nil
}

func (p *MySQLRestorePlan) UnmarshalJSON(data []byte) error {
	if len(data) > 1024*1024 || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("invalid MySQL restore JSON")
	}
	if err := pvcCopyUniqueJSON(json.NewDecoder(bytes.NewReader(data)), 0); err != nil {
		return err
	}
	fields, err := pvcCopyExactJSONKeys(data, reflect.TypeOf(MySQLRestorePlan{}))
	if err != nil {
		return err
	}
	if raw, ok := fields["items"]; ok {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		for _, item := range items {
			fields, err := pvcCopyExactJSONKeys(item, reflect.TypeOf(MySQLRestoreItem{}))
			if err != nil {
				return err
			}
			if source, ok := fields["source"]; ok {
				sourceFields, err := pvcCopyExactJSONKeys(source, reflect.TypeOf(MySQLRestoreSource{}))
				if err != nil {
					return err
				}
				if raw, ok := sourceFields["backupAdminSecretRef"]; ok {
					if _, err := pvcCopyExactJSONKeys(raw, reflect.TypeOf(MySQLRestoreCredentialRef{})); err != nil {
						return err
					}
				}
				if raw, ok := sourceFields["tls"]; ok {
					if _, err := pvcCopyExactJSONKeys(raw, reflect.TypeOf(MySQLRestoreTLS{})); err != nil {
						return err
					}
				}
			}
		}
	}
	type wire MySQLRestorePlan
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = MySQLRestorePlan(decoded)
	return nil
}
