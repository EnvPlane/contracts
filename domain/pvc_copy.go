package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

const PVCCopyContractVersion = "v1"
const RunnerOperationMaterializePVC = "materialize_pvc"
const PVCCopyMaxTimeoutSeconds int64 = 86400

type PVCCopyMethod string

const PVCCopyFilesystemOffline PVCCopyMethod = "filesystem_offline"

// PVCCopyPlan is metadata only. A digest provides integrity, not authorization.
// ValidatePVCCopyPlan must additionally receive trusted scan and policy context.
type PVCCopyPlan struct {
	ContractVersion         string        `json:"contractVersion"`
	PlanID                  string        `json:"planId"`
	TenantID                string        `json:"tenantId"`
	ProjectID               string        `json:"projectId"`
	EnvironmentID           string        `json:"environmentId"`
	TemplateRevisionID      string        `json:"templateRevisionId"`
	TemplateDigest          string        `json:"templateDigest"`
	TargetNamespace         string        `json:"targetNamespace"`
	AllowedSourceNamespaces []string      `json:"allowedSourceNamespaces"`
	Items                   []PVCCopyItem `json:"items"`
	Image                   string        `json:"image"`
	MaxBytes                int64         `json:"maxBytes"`
	TimeoutSeconds          int64         `json:"timeoutSeconds"`
	StorageQuotaBytes       int64         `json:"storageQuotaBytes"`
	Digest                  string        `json:"digest"`
}

type PVCCopyItem struct {
	ID                   string        `json:"id"`
	SourceTenantID       string        `json:"sourceTenantId"`
	SourceNamespace      string        `json:"sourceNamespace"`
	SourceName           string        `json:"sourceName"`
	SourceUID            string        `json:"sourceUid"`
	SourceIdentity       string        `json:"sourceIdentity"`
	SourceRequestedBytes int64         `json:"sourceRequestedBytes"`
	TargetName           string        `json:"targetName"`
	StorageClass         string        `json:"storageClass"`
	AccessModes          []string      `json:"accessModes"`
	RequestedBytes       int64         `json:"requestedBytes"`
	Method               PVCCopyMethod `json:"method"`
	MaxBytes             int64         `json:"maxBytes"`
	TimeoutSeconds       int64         `json:"timeoutSeconds"`
}

// PVCCopySource is projected by the server from an authenticated scan. UID must
// come from ResourceSnapshot.SourceUID (or original metadata.uid), never a
// sanitized manifest with UID removed or a user-provided replacement.
type PVCCopySource struct {
	TenantID         string   `json:"tenantId"`
	Namespace        string   `json:"namespace"`
	Name             string   `json:"name"`
	UID              string   `json:"uid"`
	StorageClass     string   `json:"storageClass"`
	AccessModes      []string `json:"accessModes"`
	RequestedBytes   int64    `json:"requestedBytes"`
	VolumeMode       string   `json:"volumeMode"`
	EnvironmentClass string   `json:"environmentClass"`
}

type PVCCopySourceRef struct {
	Namespace string
	Name      string
	UID       string
}

// PVCCopyPermission is out-of-band trusted policy, not a transport approval.
// Production approval is exact to source UID; cross-tenant copying is forbidden
// even if a source appears in ApprovedProductionSources.
type PVCCopyPermission struct {
	PlanID, TenantID, ProjectID, EnvironmentID          string
	TemplateRevisionID, TemplateDigest, TargetNamespace string
	AllowedSourceNamespaces, AllowedImages              []string
	ApprovedProductionSources                           []PVCCopySourceRef
	MaxBytes, TimeoutSeconds, StorageQuotaBytes         int64
}

var pvcCopyDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Opaque bindings may be UUIDs, digest-based IDs or slash-delimited revision IDs.
// They are never interpreted as Kubernetes names, shell arguments or file paths.
var pvcCopyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*(/[A-Za-z0-9][A-Za-z0-9._:-]*)*$`)
var pvcCopyUIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,127}$`)
var pvcCopyLabelPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
var pvcCopyImagePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9./:_-]*@sha256:[0-9a-f]{64}$`)

func pvcCopyID(s string) bool {
	return len(s) <= 512 && pvcCopyIDPattern.MatchString(s)
}

func pvcCopyName(s string, namespace bool) bool {
	if s == "" || len(s) > 253 || (namespace && len(s) > 63) {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if len(label) > 63 || !pvcCopyLabelPattern.MatchString(label) {
			return false
		}
	}
	return !namespace || !strings.Contains(s, ".")
}

func pvcCopyModes(modes []string) bool {
	if len(modes) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, mode := range modes {
		if seen[mode] || (mode != "ReadWriteOnce" && mode != "ReadWriteMany" && mode != "ReadOnlyMany" && mode != "ReadWriteOncePod") {
			return false
		}
		seen[mode] = true
	}
	return !seen["ReadWriteOncePod"] || len(modes) == 1
}

func pvcCopySameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func pvcCopyContains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func (s PVCCopySource) CanonicalIdentity() (string, error) {
	if !pvcCopyID(s.TenantID) || !pvcCopyName(s.Namespace, true) || !pvcCopyName(s.Name, false) || !pvcCopyUIDPattern.MatchString(s.UID) || !pvcCopyName(s.StorageClass, false) || !pvcCopyModes(s.AccessModes) || s.RequestedBytes <= 0 || s.VolumeMode != "Filesystem" {
		return "", errors.New("invalid filesystem PVC source metadata")
	}
	switch s.EnvironmentClass {
	case "production", "development", "staging", "test", "preview":
	default:
		return "", errors.New("PVC source environment classification is required")
	}
	s.AccessModes = append([]string(nil), s.AccessModes...)
	sort.Strings(s.AccessModes)
	return canonicalDigest(&s, func(any) {})
}

// CanonicalDigest sorts set-like fields without modifying caller-owned slices.
// Array order is not copy selection; items are bound by their unique IDs.
func (p PVCCopyPlan) CanonicalDigest() (string, error) {
	p.AllowedSourceNamespaces = append([]string(nil), p.AllowedSourceNamespaces...)
	sort.Strings(p.AllowedSourceNamespaces)
	p.Items = append([]PVCCopyItem(nil), p.Items...)
	for i := range p.Items {
		p.Items[i].AccessModes = append([]string(nil), p.Items[i].AccessModes...)
		sort.Strings(p.Items[i].AccessModes)
	}
	sort.Slice(p.Items, func(i, j int) bool { return p.Items[i].ID < p.Items[j].ID })
	return canonicalDigest(&p, func(v any) { v.(*PVCCopyPlan).Digest = "" })
}

// Validate checks structure and integrity, NOT trusted source ownership or
// approval. A consumer must also use ValidatePVCCopyPlan before execution.
func (p PVCCopyPlan) Validate() error {
	if p.ContractVersion != PVCCopyContractVersion {
		return errors.New("unsupported PVC copy contract version")
	}
	for _, id := range []string{p.PlanID, p.TenantID, p.ProjectID, p.EnvironmentID, p.TemplateRevisionID} {
		if !pvcCopyID(id) {
			return errors.New("invalid PVC copy identity binding")
		}
	}
	if !pvcCopyDigestPattern.MatchString(p.TemplateDigest) || !pvcCopyDigestPattern.MatchString(p.Digest) || !pvcCopyName(p.TargetNamespace, true) || !pvcCopyImagePattern.MatchString(p.Image) {
		return errors.New("invalid PVC copy digest, image or namespace")
	}
	if p.MaxBytes <= 0 || p.StorageQuotaBytes <= 0 || p.TimeoutSeconds <= 0 || p.TimeoutSeconds > PVCCopyMaxTimeoutSeconds || len(p.Items) == 0 {
		return errors.New("PVC copy requires positive bounded limits and items")
	}
	namespaces := map[string]bool{}
	for _, ns := range p.AllowedSourceNamespaces {
		if !pvcCopyName(ns, true) || ns == p.TargetNamespace || namespaces[ns] {
			return errors.New("invalid PVC source namespace allowlist")
		}
		namespaces[ns] = true
	}
	ids, targets, sources := map[string]bool{}, map[string]bool{}, map[string]bool{}
	var storage, copied int64
	for _, item := range p.Items {
		ref := item.SourceNamespace + "/" + item.SourceName
		if !pvcCopyID(item.ID) || ids[item.ID] || !pvcCopyName(item.TargetName, false) || targets[item.TargetName] || sources[ref] {
			return errors.New("invalid or ambiguous PVC copy selection")
		}
		ids[item.ID], targets[item.TargetName], sources[ref] = true, true, true
		if item.SourceTenantID != p.TenantID || !namespaces[item.SourceNamespace] || !pvcCopyName(item.SourceName, false) || !pvcCopyUIDPattern.MatchString(item.SourceUID) || !pvcCopyDigestPattern.MatchString(item.SourceIdentity) {
			return errors.New("PVC copy source escapes exact tenant/namespace/UID binding")
		}
		if item.Method != PVCCopyFilesystemOffline || !pvcCopyName(item.StorageClass, false) || !pvcCopyModes(item.AccessModes) || pvcCopyContains(item.AccessModes, "ReadOnlyMany") {
			return errors.New("unsupported PVC copy method or writable filesystem storage")
		}
		if item.SourceRequestedBytes <= 0 || item.RequestedBytes < item.SourceRequestedBytes || item.MaxBytes < item.SourceRequestedBytes || item.MaxBytes > item.RequestedBytes || item.TimeoutSeconds <= 0 || item.TimeoutSeconds > p.TimeoutSeconds {
			return errors.New("invalid PVC item size or timeout bounds")
		}
		// Subtraction prevents integer overflow even for malicious MaxInt64 sizes.
		if item.RequestedBytes > p.StorageQuotaBytes-storage || item.MaxBytes > p.MaxBytes-copied {
			return errors.New("PVC copy exceeds aggregate storage or copy quota")
		}
		storage += item.RequestedBytes
		copied += item.MaxBytes
	}
	actual, err := p.CanonicalDigest()
	if err != nil {
		return err
	}
	if actual != p.Digest {
		return errors.New("PVC copy plan digest mismatch")
	}
	return nil
}

// ValidatePVCCopyPlan binds the plan to trusted policy and fresh source metadata.
// The caller must derive policy from authenticated tenancy, never client flags.
func ValidatePVCCopyPlan(p PVCCopyPlan, sources []PVCCopySource, permission PVCCopyPermission) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.PlanID != permission.PlanID || p.TenantID != permission.TenantID || p.ProjectID != permission.ProjectID || p.EnvironmentID != permission.EnvironmentID || p.TemplateRevisionID != permission.TemplateRevisionID || p.TemplateDigest != permission.TemplateDigest || p.TargetNamespace != permission.TargetNamespace || !pvcCopySameSet(p.AllowedSourceNamespaces, permission.AllowedSourceNamespaces) {
		return errors.New("PVC copy permission binding mismatch")
	}
	if !pvcCopyContains(permission.AllowedImages, p.Image) || permission.MaxBytes <= 0 || p.MaxBytes > permission.MaxBytes || permission.TimeoutSeconds <= 0 || p.TimeoutSeconds > permission.TimeoutSeconds || permission.StorageQuotaBytes <= 0 || p.StorageQuotaBytes > permission.StorageQuotaBytes {
		return errors.New("PVC copy image or bounds are not permitted")
	}
	index := map[string]PVCCopySource{}
	for _, source := range sources {
		key := source.Namespace + "/" + source.Name
		if _, exists := index[key]; exists {
			return errors.New("ambiguous scanned PVC source")
		}
		index[key] = source
	}
	for _, item := range p.Items {
		source, found := index[item.SourceNamespace+"/"+item.SourceName]
		if !found {
			return fmt.Errorf("PVC source missing for item %s", item.ID)
		}
		identity, err := source.CanonicalIdentity()
		if err != nil {
			return err
		}
		if source.TenantID != p.TenantID || source.UID != item.SourceUID || identity != item.SourceIdentity || source.RequestedBytes != item.SourceRequestedBytes || source.StorageClass != item.StorageClass || !pvcCopySameSet(source.AccessModes, item.AccessModes) {
			return fmt.Errorf("PVC source identity or storage mismatch for item %s", item.ID)
		}
		if source.EnvironmentClass == "production" {
			approved := false
			for _, ref := range permission.ApprovedProductionSources {
				if ref == (PVCCopySourceRef{source.Namespace, source.Name, source.UID}) {
					approved = true
				}
			}
			if !approved {
				return errors.New("production PVC source requires exact trusted approval")
			}
		}
	}
	return nil
}

// CompilePVCCopyPlan fills only the contract version and missing source
// identities, then seals and validates. UID, selection, sizes and limits must be
// explicit; compilation never infers permission or changes the copy method.
func CompilePVCCopyPlan(p PVCCopyPlan, sources []PVCCopySource, permission PVCCopyPermission) (PVCCopyPlan, error) {
	if p.ContractVersion == "" {
		p.ContractVersion = PVCCopyContractVersion
	}
	p.AllowedSourceNamespaces = append([]string(nil), p.AllowedSourceNamespaces...)
	p.Items = append([]PVCCopyItem(nil), p.Items...)
	for i := range p.Items {
		p.Items[i].AccessModes = append([]string(nil), p.Items[i].AccessModes...)
		if p.Items[i].SourceIdentity != "" {
			continue
		}
		for _, s := range sources {
			if s.Namespace == p.Items[i].SourceNamespace && s.Name == p.Items[i].SourceName && s.UID == p.Items[i].SourceUID {
				identity, err := s.CanonicalIdentity()
				if err != nil {
					return PVCCopyPlan{}, err
				}
				p.Items[i].SourceIdentity = identity
			}
		}
	}
	digest, err := p.CanonicalDigest()
	if err != nil {
		return PVCCopyPlan{}, err
	}
	if p.Digest != "" && p.Digest != digest {
		return PVCCopyPlan{}, errors.New("PVC copy input digest mismatch")
	}
	p.Digest = digest
	if err := ValidatePVCCopyPlan(p, sources, permission); err != nil {
		return PVCCopyPlan{}, err
	}
	return p, nil
}

// UnmarshalJSON ensures unknown fields (including nested item fields) cannot be
// silently stripped before validation, including inside RunnerCommand decoding.
func (p *PVCCopyPlan) UnmarshalJSON(data []byte) error {
	if len(data) > 1024*1024 || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("invalid or oversized PVC copy plan JSON")
	}
	// Duplicate keys must not create parser-dependent authorization bindings.
	if err := pvcCopyUniqueJSON(json.NewDecoder(bytes.NewReader(data)), 0); err != nil {
		return err
	}
	fields, err := pvcCopyExactJSONKeys(data, reflect.TypeOf(PVCCopyPlan{}))
	if err != nil {
		return err
	}
	if raw, ok := fields["items"]; ok {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		for _, item := range items {
			if _, err := pvcCopyExactJSONKeys(item, reflect.TypeOf(PVCCopyItem{})); err != nil {
				return err
			}
		}
	}
	type wire PVCCopyPlan
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("PVC copy plan contains trailing JSON")
	}
	*p = PVCCopyPlan(decoded)
	return nil
}

// encoding/json otherwise accepts case-insensitive aliases of known fields.
// Require the exact wire spelling so different parsers cannot disagree.
func pvcCopyExactJSONKeys(data []byte, typ reflect.Type) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	allowed := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		allowed[strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]] = true
	}
	for key := range fields {
		if !allowed[key] {
			return nil, fmt.Errorf("unknown PVC copy JSON field %q", key)
		}
	}
	return fields, nil
}

func pvcCopyUniqueJSON(decoder *json.Decoder, depth int) error {
	if depth > 32 {
		return errors.New("PVC copy JSON nesting exceeds limit")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delimiter == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("duplicate or invalid PVC copy JSON key")
			}
			seen[name] = true
		}
		if err := pvcCopyUniqueJSON(decoder, depth+1); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
