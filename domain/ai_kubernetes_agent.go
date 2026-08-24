package domain

import (
	"errors"
	"strings"
	"time"
)

const AIKubernetesAgentSchemaVersion = "1"

type AIKubernetesDiagnosisCode string

const (
	AIKubernetesHealthy    AIKubernetesDiagnosisCode = "healthy"
	AIKubernetesCrashLoop  AIKubernetesDiagnosisCode = "crash_loop"
	AIKubernetesImagePull  AIKubernetesDiagnosisCode = "image_pull"
	AIKubernetesPendingPVC AIKubernetesDiagnosisCode = "pending_pvc"
	AIKubernetesDNS        AIKubernetesDiagnosisCode = "dns"
	AIKubernetesIngress    AIKubernetesDiagnosisCode = "ingress"
	AIKubernetesQuota      AIKubernetesDiagnosisCode = "quota"
	AIKubernetesForeign    AIKubernetesDiagnosisCode = "foreign_resource"
	AIKubernetesUnknown    AIKubernetesDiagnosisCode = "unknown"
)

type AIKubernetesResourceObservation struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
	Owned     bool   `json:"owned"`
}

type AIKubernetesEventObservation struct {
	ID         string    `json:"id"`
	Namespace  string    `json:"namespace"`
	Kind       string    `json:"kind"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Reason     string    `json:"reason"`
	ObservedAt time.Time `json:"observedAt"`
}

type AIKubernetesObservation struct {
	TenantID       string                            `json:"tenantId"`
	ProjectID      string                            `json:"projectId"`
	EnvironmentID  string                            `json:"environmentId"`
	Namespace      string                            `json:"namespace"`
	Resources      []AIKubernetesResourceObservation `json:"resources,omitempty"`
	Events         []AIKubernetesEventObservation    `json:"events,omitempty"`
	WorkloadStatus string                            `json:"workloadStatus"`
	FluxStatus     string                            `json:"fluxStatus,omitempty"`
	NetworkStatus  string                            `json:"networkStatus,omitempty"`
	StorageStatus  string                            `json:"storageStatus,omitempty"`
	QuotaStatus    string                            `json:"quotaStatus,omitempty"`
	OwnershipValid bool                              `json:"ownershipValid"`
	NamespaceValid bool                              `json:"namespaceValid"`
	Stale          bool                              `json:"stale"`
}

type AIKubernetesRepairKind string

const (
	AIKubernetesRetryRollout         AIKubernetesRepairKind = "retry_rollout"
	AIKubernetesRestartOwnedWorkload AIKubernetesRepairKind = "restart_owned_workload"
	AIKubernetesRefreshScan          AIKubernetesRepairKind = "refresh_scan"
	AIKubernetesScaleWorkload        AIKubernetesRepairKind = "scale_workload"
)

type AIKubernetesRepairProposal struct {
	Kind                 AIKubernetesRepairKind `json:"kind"`
	EnvironmentID        string                 `json:"environmentId"`
	Namespace            string                 `json:"namespace"`
	ResourceKind         string                 `json:"resourceKind,omitempty"`
	ResourceName         string                 `json:"resourceName,omitempty"`
	ApprovalRequired     bool                   `json:"approvalRequired"`
	PreVerification      []string               `json:"preVerification"`
	PostVerification     []string               `json:"postVerification"`
	CompensationGuidance string                 `json:"compensationGuidance"`
	Tool                 string                 `json:"tool"`
}

type AIKubernetesAgentPlan struct {
	SchemaVersion  string                       `json:"schemaVersion"`
	PlanID         string                       `json:"planId"`
	TenantID       string                       `json:"tenantId"`
	ProjectID      string                       `json:"projectId"`
	EnvironmentID  string                       `json:"environmentId"`
	ContextHash    string                       `json:"contextHash"`
	Observation    AIKubernetesObservation      `json:"observation"`
	Diagnosis      AIKubernetesDiagnosisCode    `json:"diagnosis"`
	Evidence       []AIEvidenceReference        `json:"evidence"`
	Proposals      []AIKubernetesRepairProposal `json:"proposals,omitempty"`
	BlockedReasons []string                     `json:"blockedReasons,omitempty"`
	FailClosed     bool                         `json:"failClosed"`
	GeneratedAt    time.Time                    `json:"generatedAt"`
}

func (o AIKubernetesObservation) Validate() error {
	if strings.TrimSpace(o.TenantID) == "" || strings.TrimSpace(o.ProjectID) == "" || strings.TrimSpace(o.EnvironmentID) == "" || strings.TrimSpace(o.Namespace) == "" || !o.NamespaceValid || !o.OwnershipValid {
		return errors.New("Kubernetes observation is outside the authorized namespace or ownership scope")
	}
	for _, value := range []string{o.TenantID, o.ProjectID, o.EnvironmentID, o.Namespace, o.WorkloadStatus, o.FluxStatus, o.NetworkStatus, o.StorageStatus, o.QuotaStatus} {
		if len(value) > 512 || strings.ContainsAny(value, "\r\n\x00") {
			return errors.New("Kubernetes observation contains an invalid value")
		}
	}
	for _, resource := range o.Resources {
		if len(resource.Name) > 256 || len(resource.Kind) > 128 || resource.Namespace != o.Namespace || !resource.Owned || strings.ContainsAny(resource.Name+resource.Kind+resource.Status+resource.Reason, "\r\n\x00") {
			return errors.New("Kubernetes resource observation is unauthorized")
		}
	}
	for _, event := range o.Events {
		if event.Namespace != o.Namespace || strings.TrimSpace(event.ID) == "" || strings.ContainsAny(event.ID+event.Kind+event.Name+event.Reason, "\r\n\x00") {
			return errors.New("Kubernetes event observation is unauthorized")
		}
	}
	return nil
}

func (p AIKubernetesAgentPlan) Validate() error {
	if p.SchemaVersion != AIKubernetesAgentSchemaVersion || strings.TrimSpace(p.PlanID) == "" || strings.TrimSpace(p.ContextHash) == "" || p.GeneratedAt.IsZero() || len(p.Evidence) == 0 {
		return errors.New("Kubernetes agent plan identity is invalid")
	}
	if p.TenantID != p.Observation.TenantID || p.ProjectID != p.Observation.ProjectID || p.EnvironmentID != p.Observation.EnvironmentID {
		return errors.New("Kubernetes agent plan scope is inconsistent")
	}
	if err := p.Observation.Validate(); err != nil {
		return err
	}
	for _, evidence := range p.Evidence {
		if evidence.TenantID != p.TenantID || strings.TrimSpace(evidence.SourceID) == "" {
			return errors.New("Kubernetes evidence is outside tenant scope")
		}
	}
	for _, proposal := range p.Proposals {
		if proposal.EnvironmentID != p.EnvironmentID || proposal.Namespace != p.Observation.Namespace || strings.TrimSpace(proposal.Tool) == "" || len(proposal.PreVerification) == 0 || len(proposal.PostVerification) == 0 || strings.TrimSpace(proposal.CompensationGuidance) == "" {
			return errors.New("Kubernetes repair proposal is invalid")
		}
		if proposal.Kind == AIKubernetesRestartOwnedWorkload || proposal.Kind == AIKubernetesScaleWorkload {
			if !proposal.ApprovalRequired {
				return errors.New("Kubernetes mutation requires approval")
			}
		}
	}
	if p.FailClosed && len(p.BlockedReasons) == 0 {
		return errors.New("fail-closed Kubernetes plan needs a reason")
	}
	return nil
}
