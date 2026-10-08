package domain

import "time"

// A failed capture records coverage, never fabricated resource identities.
type StorageCleanupAttempt struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenantId"`
	ProjectID         string    `json:"projectId"`
	EnvironmentID     string    `json:"environmentId"`
	ClusterID         string    `json:"clusterId"`
	ClusterGeneration int64     `json:"clusterGeneration"`
	ActorID           string    `json:"actorId"`
	RequestedAt       time.Time `json:"requestedAt"`
	Status            string    `json:"status"`
	Reason            string    `json:"reason"`
	SnapshotID        string    `json:"snapshotId,omitempty"`
	SecureErasure     string    `json:"secureErasure"`
}

// Storage evidence is independent of CleanupState.Verified. Backend deletion
// attestation is not forensic erasure and never grants storage mutation rights.
type StorageCleanupVolume struct {
	PVCName          string `json:"pvcName"`
	PVCUID           string `json:"pvcUid"`
	PVName           string `json:"pvName"`
	PVUID            string `json:"pvUid"`
	StorageClass     string `json:"storageClass"`
	Provisioner      string `json:"provisioner"`
	ReclaimPolicy    string `json:"reclaimPolicy"`
	BackendReference string `json:"backendReference"`
}
type StorageCleanupObservation struct {
	PVCUID           string    `json:"pvcUid"`
	PVUID            string    `json:"pvUid"`
	BackendReference string    `json:"backendReference"`
	State            string    `json:"state"`
	Source           string    `json:"source"`
	Reason           string    `json:"reason,omitempty"`
	Reference        string    `json:"reference,omitempty"`
	ObservedAt       time.Time `json:"observedAt"`
	ActorID          string    `json:"actorId,omitempty"`
	SecureErasure    string    `json:"secureErasure"`
}
type StorageCleanupSnapshot struct {
	ID                string                      `json:"id"`
	TenantID          string                      `json:"tenantId"`
	ProjectID         string                      `json:"projectId"`
	EnvironmentID     string                      `json:"environmentId"`
	ClusterID         string                      `json:"clusterId"`
	ClusterGeneration int64                       `json:"clusterGeneration"`
	Namespace         string                      `json:"namespace"`
	NamespaceUID      string                      `json:"namespaceUid"`
	Revision          int64                       `json:"revision"`
	CapturedAt        time.Time                   `json:"capturedAt"`
	Volumes           []StorageCleanupVolume      `json:"volumes"`
	Observations      []StorageCleanupObservation `json:"observations"`
}
type StorageCleanupChallenge struct {
	SnapshotID       string    `json:"snapshotId"`
	LeaseID          string    `json:"leaseId"`
	Token            string    `json:"token"`
	ExpectedRevision int64     `json:"expectedRevision"`
	ExpiresAt        time.Time `json:"expiresAt"`
}
type StorageCleanupAttestationRequest struct {
	SnapshotID       string                    `json:"snapshotId"`
	LeaseID          string                    `json:"leaseId"`
	Token            string                    `json:"token"`
	ExpectedRevision int64                     `json:"expectedRevision"`
	Observation      StorageCleanupObservation `json:"observation"`
}
