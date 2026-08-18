package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const StatefulDependencyContractVersion = "v1"

type StatefulDependencyStrategy string

const (
	StatefulStrategyEmpty            StatefulDependencyStrategy = "empty"
	StatefulStrategySeed             StatefulDependencyStrategy = "seed"
	StatefulStrategyVolumeSnapshot   StatefulDependencyStrategy = "volume_snapshot"
	StatefulStrategyDatabaseRestore  StatefulDependencyStrategy = "database_restore"
	StatefulStrategyExternalIsolated StatefulDependencyStrategy = "external_isolated_database"
	StatefulStrategyReferenceShared  StatefulDependencyStrategy = "reference_shared"
)

type StatefulDependencyPolicy struct {
	ID                   string                     `json:"id"`
	Kind                 string                     `json:"kind"`
	Required             bool                       `json:"required"`
	Strategy             StatefulDependencyStrategy `json:"strategy"`
	SourceNamespace      string                     `json:"sourceNamespace,omitempty"`
	SourceName           string                     `json:"sourceName,omitempty"`
	TargetNamespace      string                     `json:"targetNamespace"`
	TargetName           string                     `json:"targetName"`
	ServiceName          string                     `json:"serviceName,omitempty"`
	SecretRef            string                     `json:"secretRef,omitempty"`
	SeedTemplateRef      string                     `json:"seedTemplateRef,omitempty"`
	DumpRef              string                     `json:"dumpRef,omitempty"`
	RestoreCredentialRef string                     `json:"restoreCredentialRef,omitempty"`
	StorageClass         string                     `json:"storageClass,omitempty"`
	SnapshotClass        string                     `json:"snapshotClass,omitempty"`
	Size                 string                     `json:"size,omitempty"`
	AccessModes          []string                   `json:"accessModes,omitempty"`
	CSIProvisioner       string                     `json:"csiProvisioner,omitempty"`
	ApprovalRequired     bool                       `json:"approvalRequired,omitempty"`
	MaxStorage           string                     `json:"maxStorage,omitempty"`
	MaxRetries           int                        `json:"maxRetries,omitempty"`
	BackoffSeconds       int                        `json:"backoffSeconds,omitempty"`
}

type StatefulExecutionStep struct {
	ID              string `json:"id"`
	Action          string `json:"action"`
	DependencyID    string `json:"dependencyId"`
	TargetNamespace string `json:"targetNamespace"`
	TargetName      string `json:"targetName"`
	SourceNamespace string `json:"sourceNamespace,omitempty"`
	SourceName      string `json:"sourceName,omitempty"`
	TemplateRef     string `json:"templateRef,omitempty"`
	SecretRef       string `json:"secretRef,omitempty"`
	Retryable       bool   `json:"retryable"`
	MaxRetries      int    `json:"maxRetries"`
	BackoffSeconds  int    `json:"backoffSeconds"`
}

type StatefulReadinessGate struct {
	ID           string   `json:"id"`
	DependencyID string   `json:"dependencyId"`
	Conditions   []string `json:"conditions"`
	Required     bool     `json:"required"`
}

type StatefulDSNRewrite struct {
	DependencyID    string `json:"dependencyId"`
	SourceHost      string `json:"sourceHost,omitempty"`
	TargetService   string `json:"targetService"`
	TargetNamespace string `json:"targetNamespace"`
	TargetSecretRef string `json:"targetSecretRef,omitempty"`
	Path            string `json:"path"`
}

type StatefulSourceProtection struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Delete    bool   `json:"delete"`
}

type StatefulDependencyObservation struct {
	DependencyID string `json:"dependencyId"`
	State        string `json:"state"`
	Ready        bool   `json:"ready"`
	Attempt      int    `json:"attempt,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type StatefulExecutionPlan struct {
	ContractVersion    string                     `json:"contractVersion"`
	PlanID             string                     `json:"planId"`
	TenantID           string                     `json:"tenantId"`
	ProjectID          string                     `json:"projectId"`
	EnvironmentID      string                     `json:"environmentId"`
	TemplateRevisionID string                     `json:"templateRevisionId"`
	TemplateDigest     string                     `json:"templateDigest"`
	TargetNamespace    string                     `json:"targetNamespace"`
	Independent        bool                       `json:"independent"`
	Policies           []StatefulDependencyPolicy `json:"policies"`
	Steps              []StatefulExecutionStep    `json:"steps"`
	Readiness          []StatefulReadinessGate    `json:"readiness"`
	DSNRewrites        []StatefulDSNRewrite       `json:"dsnRewrites"`
	SourceProtection   []StatefulSourceProtection `json:"sourceProtection"`
	StorageQuota       string                     `json:"storageQuota,omitempty"`
	ApprovalRequired   bool                       `json:"approvalRequired"`
	InputDigest        string                     `json:"inputDigest"`
	Digest             string                     `json:"digest"`
	CreatedAt          time.Time                  `json:"createdAt"`
}

func (p StatefulExecutionPlan) CanonicalDigest() (string, error) {
	return canonicalDigest(&p, func(v any) { v.(*StatefulExecutionPlan).Digest = "" })
}

func (p StatefulExecutionPlan) Validate() error {
	if p.ContractVersion != StatefulDependencyContractVersion || p.PlanID == "" || p.TenantID == "" || p.ProjectID == "" || p.EnvironmentID == "" || p.TemplateRevisionID == "" || p.TemplateDigest == "" || p.TargetNamespace == "" || p.InputDigest == "" || p.Digest == "" {
		return errors.New("invalid stateful execution plan binding")
	}
	if len(p.Policies) == 0 {
		return errors.New("stateful execution plan has no policies")
	}
	for _, policy := range p.Policies {
		if policy.TargetNamespace != p.TargetNamespace || policy.TargetName == "" {
			return fmt.Errorf("stateful dependency %s is outside target namespace", policy.ID)
		}
		if policy.Required && policy.Strategy == "" {
			return fmt.Errorf("required stateful dependency %s has no strategy", policy.ID)
		}
		if policy.MaxRetries < 0 || policy.BackoffSeconds < 0 {
			return fmt.Errorf("invalid retry policy for %s", policy.ID)
		}
		switch policy.Strategy {
		case StatefulStrategyEmpty:
		case StatefulStrategySeed:
			if policy.SeedTemplateRef == "" {
				return fmt.Errorf("seed template is required for %s", policy.ID)
			}
		case StatefulStrategyVolumeSnapshot:
			if policy.SourceNamespace == "" || policy.SourceName == "" || policy.StorageClass == "" || policy.SnapshotClass == "" || policy.Size == "" || len(policy.AccessModes) == 0 || policy.CSIProvisioner == "" {
				return fmt.Errorf("snapshot capability and source/storage settings are required for %s", policy.ID)
			}
		case StatefulStrategyDatabaseRestore:
			if policy.DumpRef == "" || policy.RestoreCredentialRef == "" {
				return fmt.Errorf("restore refs are required for %s", policy.ID)
			}
		case StatefulStrategyExternalIsolated:
			if policy.ServiceName == "" || policy.SecretRef == "" {
				return fmt.Errorf("isolated database service and secret refs are required for %s", policy.ID)
			}
		case StatefulStrategyReferenceShared:
			if policy.SourceNamespace == "" || policy.SourceName == "" {
				return fmt.Errorf("shared reference source is required for %s", policy.ID)
			}
		default:
			return fmt.Errorf("unsupported stateful strategy %q", policy.Strategy)
		}
	}
	if p.Independent && hasSharedReference(p.Policies) {
		return errors.New("shared stateful reference cannot be marked independent")
	}
	actual, err := p.CanonicalDigest()
	if err != nil {
		return err
	}
	if actual != p.Digest {
		return fmt.Errorf("stateful plan digest mismatch: got %s want %s", p.Digest, actual)
	}
	return nil
}

func CompileStatefulExecutionPlan(tenantID, projectID, environmentID, revisionID, revisionDigest, targetNamespace string, policies []StatefulDependencyPolicy, inputDigest string, now time.Time) (StatefulExecutionPlan, error) {
	if tenantID == "" || projectID == "" || environmentID == "" || revisionID == "" || revisionDigest == "" || targetNamespace == "" {
		return StatefulExecutionPlan{}, errors.New("stateful plan identity is incomplete")
	}
	plan := StatefulExecutionPlan{ContractVersion: StatefulDependencyContractVersion, PlanID: revisionID + "/" + environmentID + "/stateful", TenantID: tenantID, ProjectID: projectID, EnvironmentID: environmentID, TemplateRevisionID: revisionID, TemplateDigest: revisionDigest, TargetNamespace: targetNamespace, Policies: append([]StatefulDependencyPolicy(nil), policies...), Independent: true, InputDigest: inputDigest, CreatedAt: now.UTC()}
	for i := range plan.Policies {
		policy := &plan.Policies[i]
		if policy.TargetNamespace == "" {
			policy.TargetNamespace = targetNamespace
		}
		if policy.TargetNamespace != targetNamespace {
			return StatefulExecutionPlan{}, fmt.Errorf("stateful dependency %s escapes target namespace", policy.ID)
		}
		if policy.Required && policy.Strategy == "" {
			return StatefulExecutionPlan{}, fmt.Errorf("required stateful dependency %s has no strategy", policy.ID)
		}
		if policy.MaxRetries == 0 {
			policy.MaxRetries = 3
		}
		if policy.BackoffSeconds == 0 {
			policy.BackoffSeconds = 5
		}
		if policy.Strategy == StatefulStrategyReferenceShared {
			plan.Independent = false
		}
		if policy.ApprovalRequired || (policy.MaxStorage != "" && policy.Size == policy.MaxStorage) {
			plan.ApprovalRequired = true
		}
		step := StatefulExecutionStep{ID: policy.ID + "-prepare", Action: "create_empty", DependencyID: policy.ID, TargetNamespace: policy.TargetNamespace, TargetName: policy.TargetName, Retryable: true, MaxRetries: policy.MaxRetries, BackoffSeconds: policy.BackoffSeconds}
		switch policy.Strategy {
		case StatefulStrategySeed:
			step.Action = "run_seed"
			step.TemplateRef = policy.SeedTemplateRef
		case StatefulStrategyVolumeSnapshot:
			step.Action = "clone_volume_snapshot"
			step.SourceNamespace = policy.SourceNamespace
			step.SourceName = policy.SourceName
		case StatefulStrategyDatabaseRestore:
			step.Action = "restore_database_dump"
			step.TemplateRef = policy.DumpRef
			step.SecretRef = policy.RestoreCredentialRef
		case StatefulStrategyExternalIsolated:
			step.Action = "bind_external_database"
			step.SecretRef = policy.SecretRef
		case StatefulStrategyReferenceShared:
			step.Action = "bind_shared_reference"
			plan.Independent = false
		}
		plan.Steps = append(plan.Steps, step)
		plan.Readiness = append(plan.Readiness, StatefulReadinessGate{ID: policy.ID + "-ready", DependencyID: policy.ID, Conditions: []string{"service_available", "secret_available", "storage_bound", "seed_or_restore_succeeded"}, Required: policy.Required})
		if policy.ServiceName != "" || policy.SecretRef != "" {
			plan.DSNRewrites = append(plan.DSNRewrites, StatefulDSNRewrite{DependencyID: policy.ID, SourceHost: policy.SourceName, TargetService: firstNonEmptyStateful(policy.ServiceName, policy.TargetName), TargetNamespace: policy.TargetNamespace, TargetSecretRef: policy.SecretRef, Path: policy.ID})
		}
		if policy.SourceName != "" {
			plan.SourceProtection = append(plan.SourceProtection, StatefulSourceProtection{Kind: policy.Kind, Namespace: policy.SourceNamespace, Name: policy.SourceName, Delete: false})
		}
	}
	sort.Slice(plan.Policies, func(i, j int) bool { return plan.Policies[i].ID < plan.Policies[j].ID })
	sort.Slice(plan.Steps, func(i, j int) bool { return plan.Steps[i].ID < plan.Steps[j].ID })
	sort.Slice(plan.Readiness, func(i, j int) bool { return plan.Readiness[i].ID < plan.Readiness[j].ID })
	sort.Slice(plan.DSNRewrites, func(i, j int) bool { return plan.DSNRewrites[i].DependencyID < plan.DSNRewrites[j].DependencyID })
	digest, err := plan.CanonicalDigest()
	if err != nil {
		return StatefulExecutionPlan{}, err
	}
	plan.Digest = digest
	return plan, nil
}

func (p StatefulExecutionPlan) CanDeleteSource(kind, namespace, name string) bool {
	for _, item := range p.SourceProtection {
		if item.Kind == kind && item.Namespace == namespace && item.Name == name {
			return item.Delete
		}
	}
	return false
}
func (p StatefulExecutionPlan) CanDeleteTarget(kind, namespace, name string) bool {
	if namespace != p.TargetNamespace || p.CanDeleteSource(kind, namespace, name) {
		return false
	}
	for _, step := range p.Steps {
		if step.TargetNamespace == namespace && step.TargetName == name && step.Action != "bind_shared_reference" && step.Action != "bind_external_database" {
			return true
		}
	}
	return false
}
func EnforceStatefulQuota(plan StatefulExecutionPlan, quota string, approvalGranted bool) error {
	if quota == "" {
		return nil
	}
	if plan.ApprovalRequired && !approvalGranted {
		return errors.New("stateful clone requires explicit approval")
	}
	total := int64(0)
	for _, policy := range plan.Policies {
		if policy.Size != "" {
			value, err := storageBytes(policy.Size)
			if err != nil {
				return err
			}
			total += value
		}
	}
	limit, err := storageBytes(quota)
	if err != nil {
		return err
	}
	if total > limit {
		return fmt.Errorf("stateful storage quota exceeded: requested %s, limit %s", formatStorage(total), quota)
	}
	return nil
}
func hasSharedReference(policies []StatefulDependencyPolicy) bool {
	for _, policy := range policies {
		if policy.Strategy == StatefulStrategyReferenceShared {
			return true
		}
	}
	return false
}
func firstNonEmptyStateful(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
func storageBytes(raw string) (int64, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	units := []struct {
		suffix     string
		multiplier int64
	}{{"gi", 1024 * 1024 * 1024}, {"mi", 1024 * 1024}, {"ki", 1024}, {"b", 1}}
	for _, unit := range units {
		if strings.HasSuffix(value, unit.suffix) {
			var number int64
			if _, err := fmt.Sscanf(strings.TrimSuffix(value, unit.suffix), "%d", &number); err != nil || number < 0 {
				break
			}
			return number * unit.multiplier, nil
		}
	}
	return 0, fmt.Errorf("invalid storage quantity %q", raw)
}
func formatStorage(value int64) string { return fmt.Sprintf("%dB", value) }
