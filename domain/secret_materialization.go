package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const SecretMaterializationContractVersion = "v1"

const DefaultSecretRetentionHours = 168

type SecretMaterializationPlanState string

const (
	SecretPlanPending       SecretMaterializationPlanState = "pending"
	SecretPlanApproved      SecretMaterializationPlanState = "approved"
	SecretPlanMaterializing SecretMaterializationPlanState = "materializing"
	SecretPlanReady         SecretMaterializationPlanState = "ready"
	SecretPlanFailed        SecretMaterializationPlanState = "failed"
	SecretPlanCleaning      SecretMaterializationPlanState = "cleaning"
	SecretPlanDeleted       SecretMaterializationPlanState = "deleted"
)

type SecretMaterializationItemState string

const (
	SecretItemPending       SecretMaterializationItemState = "pending"
	SecretItemMaterializing SecretMaterializationItemState = "materializing"
	SecretItemReady         SecretMaterializationItemState = "ready"
	SecretItemFailed        SecretMaterializationItemState = "failed"
	SecretItemCleaning      SecretMaterializationItemState = "cleaning"
	SecretItemDeleted       SecretMaterializationItemState = "deleted"
)

type SecretMaterializationOperation string

const (
	SecretOperationMaterialize SecretMaterializationOperation = "materialize"
	SecretOperationCleanup     SecretMaterializationOperation = "cleanup"
)

type SecretMaterializationErrorCode string

const (
	SecretErrorInvalidBinding      SecretMaterializationErrorCode = "invalid_binding"
	SecretErrorNamespaceEscape     SecretMaterializationErrorCode = "namespace_escape"
	SecretErrorMissingBinding      SecretMaterializationErrorCode = "missing_binding"
	SecretErrorPlaintextForbidden  SecretMaterializationErrorCode = "plaintext_forbidden"
	SecretErrorUnsupportedStrategy SecretMaterializationErrorCode = "unsupported_strategy"
	SecretErrorAmbiguousOwnership  SecretMaterializationErrorCode = "ambiguous_ownership"
	SecretErrorConflict            SecretMaterializationErrorCode = "conflict"
	SecretErrorSourceNotFound      SecretMaterializationErrorCode = "source_not_found"
	SecretErrorBackendUnavailable  SecretMaterializationErrorCode = "backend_unavailable"
	SecretErrorTimeout             SecretMaterializationErrorCode = "timeout"
	SecretErrorPermissionDenied    SecretMaterializationErrorCode = "permission_denied"
	SecretErrorValidationFailed    SecretMaterializationErrorCode = "validation_failed"
)

func (e SecretMaterializationErrorCode) Retryable() bool {
	switch e {
	case SecretErrorConflict, SecretErrorBackendUnavailable, SecretErrorTimeout:
		return true
	default:
		return false
	}
}

func (s SecretMaterializationStrategy) Action(operation SecretMaterializationOperation) string {
	if operation == SecretOperationCleanup {
		return "delete_owned_secret"
	}
	switch s {
	case SecretStrategyReference:
		return "bind_existing_secret"
	case SecretStrategyExternal:
		return "resolve_external_secret"
	case SecretStrategyEncryptedClone:
		return "decrypt_and_clone_secret"
	case SecretStrategyManual:
		return "await_manual_secret_reference"
	case SecretStrategyGenerated:
		return "generate_secret"
	default:
		return ""
	}
}

type SecretMaterializationItemResult struct {
	ItemID          string                         `json:"itemId"`
	Strategy        SecretMaterializationStrategy  `json:"strategy"`
	TargetNamespace string                         `json:"targetNamespace"`
	TargetName      string                         `json:"targetName"`
	Operation       SecretMaterializationOperation `json:"operation"`
	IdempotencyKey  string                         `json:"idempotencyKey"`
	InputDigest     string                         `json:"inputDigest"`
	OutputDigest    string                         `json:"outputDigest,omitempty"`
	Status          SecretMaterializationItemState `json:"status"`
	ErrorCode       SecretMaterializationErrorCode `json:"errorCode,omitempty"`
	Attempt         int                            `json:"attempt"`
	StartedAt       time.Time                      `json:"startedAt"`
	FinishedAt      time.Time                      `json:"finishedAt,omitempty"`
}

type SecretMaterializationConcurrency struct {
	ExpectedRevision int64 `json:"expectedRevision"`
	CurrentRevision  int64 `json:"currentRevision"`
}

type SecretMaterializationStrategy string

const (
	SecretStrategyReference      SecretMaterializationStrategy = "reference"
	SecretStrategyExternal       SecretMaterializationStrategy = "external"
	SecretStrategyEncryptedClone SecretMaterializationStrategy = "encrypted_clone"
	SecretStrategyManual         SecretMaterializationStrategy = "manual"
	SecretStrategyGenerated      SecretMaterializationStrategy = "generated"
)

// SecretStrategyConfig is the persisted, non-secret part of a UI strategy.
// ManualValue is write-only and deliberately omitted from every JSON form.
type SecretStrategyConfig struct {
	ID                  string                        `json:"id"`
	Strategy            SecretMaterializationStrategy `json:"strategy"`
	Required            bool                          `json:"required"`
	SourceNamespace     string                        `json:"sourceNamespace,omitempty"`
	SourceTenantID      string                        `json:"sourceTenantId,omitempty"`
	SourceName          string                        `json:"sourceName,omitempty"`
	TargetNamespace     string                        `json:"targetNamespace"`
	TargetName          string                        `json:"targetName"`
	Backend             string                        `json:"backend,omitempty"`
	ExternalSecretStore string                        `json:"externalSecretStore,omitempty"`
	ExternalKey         string                        `json:"externalKey,omitempty"`
	EncryptedPayloadRef string                        `json:"encryptedPayloadRef,omitempty"`
	Generator           string                        `json:"generator,omitempty"`
	CredentialRotation  string                        `json:"credentialRotation,omitempty"`
	RetentionHours      int                           `json:"retentionHours,omitempty"`
	ApprovalRequired    bool                          `json:"approvalRequired,omitempty"`
	ManualValue         string                        `json:"-"`
}

type SecretMaterializationItem struct {
	ID                  string                        `json:"id"`
	Strategy            SecretMaterializationStrategy `json:"strategy"`
	Required            bool                          `json:"required"`
	SourceNamespace     string                        `json:"sourceNamespace,omitempty"`
	SourceTenantID      string                        `json:"sourceTenantId,omitempty"`
	SourceName          string                        `json:"sourceName,omitempty"`
	TargetNamespace     string                        `json:"targetNamespace"`
	TargetName          string                        `json:"targetName"`
	Backend             string                        `json:"backend,omitempty"`
	ExternalSecretStore string                        `json:"externalSecretStore,omitempty"`
	ExternalKey         string                        `json:"externalKey,omitempty"`
	EncryptedPayloadRef string                        `json:"encryptedPayloadRef,omitempty"`
	Generator           string                        `json:"generator,omitempty"`
	CredentialRotation  string                        `json:"credentialRotation,omitempty"`
	RetentionHours      int                           `json:"retentionHours,omitempty"`
	Owned               bool                          `json:"owned"`
}

type SecretMaterializationPlan struct {
	ContractVersion         string                            `json:"contractVersion"`
	PlanID                  string                            `json:"planId"`
	TenantID                string                            `json:"tenantId"`
	ProjectID               string                            `json:"projectId"`
	EnvironmentID           string                            `json:"environmentId"`
	TemplateRevisionID      string                            `json:"templateRevisionId"`
	TemplateDigest          string                            `json:"templateDigest"`
	TargetNamespace         string                            `json:"targetNamespace"`
	AllowedSourceNamespaces []string                          `json:"allowedSourceNamespaces,omitempty"`
	Items                   []SecretMaterializationItem       `json:"items"`
	Ownership               []OwnershipRecord                 `json:"ownership"`
	ApprovalRequired        bool                              `json:"approvalRequired,omitempty"`
	InputDigest             string                            `json:"inputDigest"`
	Digest                  string                            `json:"digest"`
	CreatedAt               time.Time                         `json:"createdAt"`
	State                   SecretMaterializationPlanState    `json:"state,omitempty"`
	Revision                int64                             `json:"revision,omitempty"`
	Results                 []SecretMaterializationItemResult `json:"results,omitempty"`
}

func (p SecretMaterializationPlan) Validate() error {
	if p.ContractVersion != SecretMaterializationContractVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.EnvironmentID) == "" || strings.TrimSpace(p.TemplateRevisionID) == "" || strings.TrimSpace(p.TemplateDigest) == "" || strings.TrimSpace(p.TargetNamespace) == "" || strings.TrimSpace(p.InputDigest) == "" || strings.TrimSpace(p.Digest) == "" {
		return errors.New("invalid secret materialization plan binding")
	}
	if p.State != "" && !validSecretPlanState(p.State) {
		return fmt.Errorf("unsupported secret materialization plan state %q", p.State)
	}
	actualDigest, err := p.CanonicalDigest()
	if err != nil {
		return err
	}
	if actualDigest != p.Digest {
		return fmt.Errorf("secret materialization plan digest mismatch: got %s want %s", p.Digest, actualDigest)
	}
	owned := map[string]bool{}
	itemIDs := map[string]bool{}
	targets := map[string]bool{}
	ownershipSeen := map[string]bool{}
	itemsByID := map[string]SecretMaterializationItem{}
	for _, item := range p.Items {
		if strings.TrimSpace(item.ID) == "" || itemIDs[item.ID] || strings.TrimSpace(item.TargetNamespace) != p.TargetNamespace || !validSecretNamespace(item.TargetNamespace) || strings.TrimSpace(item.TargetName) == "" {
			if itemIDs[item.ID] {
				return fmt.Errorf("duplicate secret materialization item %q", item.ID)
			}
			return errors.New("secret materialization target is outside the exact target namespace")
		}
		itemIDs[item.ID] = true
		itemsByID[item.ID] = item
		if targets[item.TargetName] {
			return fmt.Errorf("ambiguous secret ownership for target %q", item.TargetName)
		}
		targets[item.TargetName] = true
		if item.RetentionHours <= 0 {
			return fmt.Errorf("secret %s has no bounded retention", item.ID)
		}
		if item.SourceTenantID != "" && item.SourceTenantID != p.TenantID && (!p.ApprovalRequired || p.State != SecretPlanApproved) {
			return fmt.Errorf("cross-tenant secret %s requires approval", item.ID)
		}
		switch item.Strategy {
		case SecretStrategyReference:
			if item.SourceNamespace == "" || item.SourceName == "" || item.Owned || !containsString(p.AllowedSourceNamespaces, item.SourceNamespace) {
				return errors.New("reference strategy source is not explicitly allowlisted")
			}
		case SecretStrategyExternal:
			if item.ExternalSecretStore == "" || item.ExternalKey == "" || item.Owned {
				return errors.New("external strategy requires a non-owned store reference")
			}
		case SecretStrategyEncryptedClone, SecretStrategyManual:
			if item.EncryptedPayloadRef == "" || !item.Owned {
				return errors.New("encrypted strategy requires an owned encrypted payload reference")
			}
		case SecretStrategyGenerated:
			if item.Generator == "" || item.CredentialRotation == "" || !item.Owned {
				return errors.New("generated strategy requires an owned generator")
			}
		default:
			return fmt.Errorf("unsupported secret materialization strategy %q", item.Strategy)
		}
		if item.Owned {
			owned[item.TargetName] = true
		}
	}
	for _, record := range p.Ownership {
		if record.Kind != "Secret" || record.Namespace != p.TargetNamespace || !validSecretNamespace(record.Namespace) || record.Name == "" {
			return errors.New("secret ownership is not exact and namespaced")
		}
		if ownershipSeen[record.Name] {
			return fmt.Errorf("ambiguous secret ownership for target %q", record.Name)
		}
		ownershipSeen[record.Name] = true
	}
	for _, record := range p.Ownership {
		if !owned[record.Name] {
			return errors.New("ownership contains a non-owned secret")
		}
	}
	for _, result := range p.Results {
		item, itemExists := itemsByID[result.ItemID]
		if !itemExists || result.Strategy != item.Strategy || result.TargetNamespace != item.TargetNamespace || result.TargetName != item.TargetName || !validSecretItemState(result.Status) || result.Attempt < 1 || result.IdempotencyKey == "" || result.InputDigest == "" || result.StartedAt.IsZero() {
			return errors.New("invalid secret materialization item result")
		}
		expectedKey, err := SecretMaterializationIdempotencyKey(p.TenantID, p.ProjectID, p.EnvironmentID, p.TemplateDigest, p.TargetNamespace, result.ItemID, result.Operation)
		if err != nil || result.IdempotencyKey != expectedKey {
			return fmt.Errorf("item %s has an invalid idempotency key", result.ItemID)
		}
		if result.Operation != SecretOperationMaterialize && result.Operation != SecretOperationCleanup {
			return fmt.Errorf("unsupported secret materialization operation %q", result.Operation)
		}
		if result.Operation == SecretOperationCleanup && !item.Owned {
			return fmt.Errorf("item %s cannot clean up a non-owned secret", result.ItemID)
		}
		if result.ErrorCode != "" && result.Status != SecretItemFailed {
			return errors.New("error code is only valid for failed secret item results")
		}
		if result.ErrorCode != "" && !validSecretErrorCode(result.ErrorCode) {
			return fmt.Errorf("unsupported secret materialization error code %q", result.ErrorCode)
		}
		if result.OutputDigest != "" && result.Status != SecretItemReady && result.Status != SecretItemDeleted {
			return errors.New("output digest is only valid for ready or deleted results")
		}
	}
	return nil
}

func (p SecretMaterializationPlan) CanDeleteOwnedSecret(namespace, name string) bool {
	if namespace != p.TargetNamespace {
		return false
	}
	for _, record := range p.Ownership {
		if record.Kind == "Secret" && record.Namespace == namespace && record.Name == name {
			return true
		}
	}
	return false
}

// CompileSecretMaterializationPlan converts saved strategy metadata into a
// runtime plan. It never accepts or copies plaintext and never serializes the
// encrypted envelope; the reference is resolved only at execution time.
func CompileSecretMaterializationPlan(tenantID, projectID, environmentID, revisionID, revisionDigest, targetNamespace string, configs []SecretStrategyConfig, inputDigest string, now time.Time) (SecretMaterializationPlan, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(projectID) == "" || strings.TrimSpace(environmentID) == "" || strings.TrimSpace(revisionID) == "" || strings.TrimSpace(revisionDigest) == "" || strings.TrimSpace(targetNamespace) == "" {
		return SecretMaterializationPlan{}, errors.New("secret plan identity is incomplete")
	}
	items := make([]SecretMaterializationItem, 0, len(configs))
	ownership := make([]OwnershipRecord, 0)
	allowlist := make([]string, 0)
	approvalRequired := false
	for _, cfg := range configs {
		if strings.TrimSpace(cfg.TargetNamespace) == "" {
			cfg.TargetNamespace = targetNamespace
		}
		if cfg.TargetNamespace != targetNamespace {
			return SecretMaterializationPlan{}, errors.New("secret strategy namespace is outside the environment namespace")
		}
		if cfg.ManualValue != "" {
			return SecretMaterializationPlan{}, errors.New("plaintext secret value is not accepted by plan compiler")
		}
		item := SecretMaterializationItem{ID: cfg.ID, Strategy: cfg.Strategy, Required: cfg.Required, SourceNamespace: cfg.SourceNamespace, SourceName: cfg.SourceName, TargetNamespace: cfg.TargetNamespace, TargetName: cfg.TargetName, Backend: cfg.Backend, ExternalSecretStore: cfg.ExternalSecretStore, ExternalKey: cfg.ExternalKey, EncryptedPayloadRef: cfg.EncryptedPayloadRef, Generator: cfg.Generator}
		item.SourceTenantID = cfg.SourceTenantID
		item.CredentialRotation = cfg.CredentialRotation
		item.RetentionHours = cfg.RetentionHours
		if item.RetentionHours == 0 {
			item.RetentionHours = DefaultSecretRetentionHours
		}
		if item.Strategy == SecretStrategyGenerated && item.CredentialRotation == "" {
			item.CredentialRotation = "on_create_and_cleanup"
		}
		if item.Strategy == SecretStrategyReference && item.SourceNamespace != "" && !containsString(allowlist, item.SourceNamespace) {
			allowlist = append(allowlist, item.SourceNamespace)
		}
		if cfg.ApprovalRequired {
			// Approval is represented by the plan state; this flag records that
			// the strategy cannot proceed until an approval is recorded.
			approvalRequired = true
		}
		item.Owned = cfg.Strategy == SecretStrategyEncryptedClone || cfg.Strategy == SecretStrategyManual || cfg.Strategy == SecretStrategyGenerated
		if item.Owned {
			ownership = append(ownership, OwnershipRecord{Kind: "Secret", Namespace: item.TargetNamespace, Name: item.TargetName})
		}
		items = append(items, item)
	}
	plan := SecretMaterializationPlan{ContractVersion: SecretMaterializationContractVersion, PlanID: revisionID + "/" + environmentID + "/secrets", TenantID: tenantID, ProjectID: projectID, EnvironmentID: environmentID, TemplateRevisionID: revisionID, TemplateDigest: revisionDigest, TargetNamespace: targetNamespace, AllowedSourceNamespaces: allowlist, Items: items, Ownership: ownership, ApprovalRequired: approvalRequired, InputDigest: inputDigest, CreatedAt: now.UTC(), State: SecretPlanPending, Revision: 1}
	for _, item := range items {
		if item.SourceTenantID != "" && item.SourceTenantID != tenantID {
			plan.ApprovalRequired = true
		}
	}
	digest, err := plan.CanonicalDigest()
	if err != nil {
		return SecretMaterializationPlan{}, err
	}
	plan.Digest = digest
	return plan, nil
}

func (p SecretMaterializationPlan) CanonicalDigest() (string, error) {
	return canonicalDigest(&p, func(v any) {
		plan := v.(*SecretMaterializationPlan)
		plan.Digest = ""
		plan.State = ""
		plan.Revision = 0
		plan.Results = nil
	})
}

// SecretMaterializationIdempotencyKey is stable across retries and independent
// of map ordering. It intentionally contains no secret value or payload.
func SecretMaterializationIdempotencyKey(tenantID, projectID, environmentID, templateDigest, targetNamespace, itemID string, operation SecretMaterializationOperation) (string, error) {
	values := struct {
		TenantID        string                         `json:"tenantId"`
		ProjectID       string                         `json:"projectId"`
		EnvironmentID   string                         `json:"environmentId"`
		TemplateDigest  string                         `json:"templateDigest"`
		TargetNamespace string                         `json:"targetNamespace"`
		ItemID          string                         `json:"itemId"`
		Operation       SecretMaterializationOperation `json:"operation"`
	}{tenantID, projectID, environmentID, templateDigest, targetNamespace, itemID, operation}
	for _, value := range []string{tenantID, projectID, environmentID, templateDigest, targetNamespace, itemID, string(operation)} {
		if strings.TrimSpace(value) == "" {
			return "", errors.New("secret materialization idempotency binding is incomplete")
		}
	}
	return canonicalDigest(&values, func(any) {})
}

func ValidateSecretMaterializationRevision(expected, current int64) error {
	if expected != current {
		return fmt.Errorf("secret materialization revision conflict: expected %d, current %d", expected, current)
	}
	return nil
}

func ValidateSecretMaterializationTransition(from, to SecretMaterializationPlanState) error {
	if !validSecretPlanState(from) || !validSecretPlanState(to) {
		return errors.New("unsupported secret materialization state transition")
	}
	if from == to {
		return nil
	}
	valid := map[SecretMaterializationPlanState][]SecretMaterializationPlanState{
		SecretPlanPending:       {SecretPlanApproved, SecretPlanFailed},
		SecretPlanApproved:      {SecretPlanMaterializing, SecretPlanFailed},
		SecretPlanMaterializing: {SecretPlanReady, SecretPlanFailed},
		SecretPlanReady:         {SecretPlanCleaning, SecretPlanFailed},
		SecretPlanFailed:        {SecretPlanMaterializing, SecretPlanCleaning},
		SecretPlanCleaning:      {SecretPlanDeleted, SecretPlanFailed},
	}
	for _, candidate := range valid[from] {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("invalid secret materialization state transition %q to %q", from, to)
}

func validSecretPlanState(state SecretMaterializationPlanState) bool {
	switch state {
	case SecretPlanPending, SecretPlanApproved, SecretPlanMaterializing, SecretPlanReady, SecretPlanFailed, SecretPlanCleaning, SecretPlanDeleted:
		return true
	}
	return false
}
func validSecretItemState(state SecretMaterializationItemState) bool {
	switch state {
	case SecretItemPending, SecretItemMaterializing, SecretItemReady, SecretItemFailed, SecretItemCleaning, SecretItemDeleted:
		return true
	}
	return false
}
func validSecretErrorCode(code SecretMaterializationErrorCode) bool {
	switch code {
	case SecretErrorInvalidBinding, SecretErrorNamespaceEscape, SecretErrorMissingBinding, SecretErrorPlaintextForbidden, SecretErrorUnsupportedStrategy, SecretErrorAmbiguousOwnership, SecretErrorConflict, SecretErrorSourceNotFound, SecretErrorBackendUnavailable, SecretErrorTimeout, SecretErrorPermissionDenied, SecretErrorValidationFailed:
		return true
	}
	return false
}

func validSecretNamespace(namespace string) bool {
	namespace = strings.TrimSpace(namespace)
	return namespace != "" && namespace != "." && namespace != ".." && !strings.ContainsAny(namespace, "/\\")
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
