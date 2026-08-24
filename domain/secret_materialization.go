package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const SecretMaterializationContractVersion = "v1"

const DefaultSecretRetentionHours = 168

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
	ContractVersion         string                      `json:"contractVersion"`
	PlanID                  string                      `json:"planId"`
	TenantID                string                      `json:"tenantId"`
	ProjectID               string                      `json:"projectId"`
	EnvironmentID           string                      `json:"environmentId"`
	TemplateRevisionID      string                      `json:"templateRevisionId"`
	TemplateDigest          string                      `json:"templateDigest"`
	TargetNamespace         string                      `json:"targetNamespace"`
	AllowedSourceNamespaces []string                    `json:"allowedSourceNamespaces,omitempty"`
	Items                   []SecretMaterializationItem `json:"items"`
	Ownership               []OwnershipRecord           `json:"ownership"`
	ApprovalRequired        bool                        `json:"approvalRequired,omitempty"`
	InputDigest             string                      `json:"inputDigest"`
	Digest                  string                      `json:"digest"`
	CreatedAt               time.Time                   `json:"createdAt"`
}

func (p SecretMaterializationPlan) Validate() error {
	if p.ContractVersion != SecretMaterializationContractVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.EnvironmentID) == "" || strings.TrimSpace(p.TemplateRevisionID) == "" || strings.TrimSpace(p.TemplateDigest) == "" || strings.TrimSpace(p.TargetNamespace) == "" || strings.TrimSpace(p.InputDigest) == "" || strings.TrimSpace(p.Digest) == "" {
		return errors.New("invalid secret materialization plan binding")
	}
	owned := map[string]bool{}
	for _, item := range p.Items {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.TargetNamespace) != p.TargetNamespace || strings.TrimSpace(item.TargetName) == "" {
			return errors.New("secret materialization target is outside the exact target namespace")
		}
		if item.RetentionHours <= 0 {
			return fmt.Errorf("secret %s has no bounded retention", item.ID)
		}
		if item.SourceTenantID != "" && item.SourceTenantID != p.TenantID && !p.ApprovalRequired {
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
		if record.Kind != "Secret" || record.Namespace != p.TargetNamespace || record.Name == "" {
			return errors.New("secret ownership is not exact and namespaced")
		}
	}
	for _, record := range p.Ownership {
		if !owned[record.Name] {
			return errors.New("ownership contains a non-owned secret")
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
		item.Owned = cfg.Strategy == SecretStrategyEncryptedClone || cfg.Strategy == SecretStrategyManual || cfg.Strategy == SecretStrategyGenerated
		if item.Owned {
			ownership = append(ownership, OwnershipRecord{Kind: "Secret", Namespace: item.TargetNamespace, Name: item.TargetName})
		}
		items = append(items, item)
	}
	plan := SecretMaterializationPlan{ContractVersion: SecretMaterializationContractVersion, PlanID: revisionID + "/" + environmentID + "/secrets", TenantID: tenantID, ProjectID: projectID, EnvironmentID: environmentID, TemplateRevisionID: revisionID, TemplateDigest: revisionDigest, TargetNamespace: targetNamespace, AllowedSourceNamespaces: allowlist, Items: items, Ownership: ownership, InputDigest: inputDigest, CreatedAt: now.UTC()}
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
	return canonicalDigest(&p, func(v any) { v.(*SecretMaterializationPlan).Digest = "" })
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
