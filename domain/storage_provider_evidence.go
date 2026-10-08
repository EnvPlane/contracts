package domain

import "time"

// Different signing domains prevent a provider receipt from being substituted
// for an independent erasure statement, even if an operator misroutes a payload.
const StorageProviderReportSigningDomain = "envplane.storage.provider.report.v1\x00"
const StorageErasureStatementSigningDomain = "envplane.storage.erasure.statement.v1\x00"

// The verifier receives opaque captured identities, never a caller-chosen path
// or an instruction to erase storage. Nonces are transient and must not be logged.
type StorageProviderChallenge struct {
	Version           string               `json:"version"`
	ID                string               `json:"id"`
	Nonce             string               `json:"nonce"`
	SnapshotID        string               `json:"snapshotId"`
	ExpectedRevision  int64                `json:"expectedRevision"`
	TenantID          string               `json:"tenantId"`
	ProjectID         string               `json:"projectId"`
	EnvironmentID     string               `json:"environmentId"`
	ClusterID         string               `json:"clusterId"`
	ClusterGeneration int64                `json:"clusterGeneration"`
	NamespaceUID      string               `json:"namespaceUid"`
	CapturedAt        time.Time            `json:"capturedAt"`
	Volume            StorageCleanupVolume `json:"volume"`
	IssuedAt          time.Time            `json:"issuedAt"`
	ExpiresAt         time.Time            `json:"expiresAt"`
}

type StorageProviderReport struct {
	Challenge          StorageProviderChallenge   `json:"challenge"`
	State              string                     `json:"state"`
	Reason             string                     `json:"reason"`
	ObservedAt         time.Time                  `json:"observedAt"`
	ErasureCertificate *StorageErasureCertificate `json:"erasureCertificate,omitempty"`
}

// Payloads are canonical JSON encoded using unpadded base64url. A receipt is
// accepted only against deployment-pinned public keys, not client-provided keys.
type StorageSignedProviderReport struct {
	KeyID     string `json:"keyId"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

type StorageErasureCertificate struct {
	KeyID     string `json:"keyId"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

// A trusted certificate covers this captured volume only. Neither provider
// deletion nor cryptographic erasure implicitly covers snapshots, backups,
// exports, rollback copies or media outside this exact identity.
type StorageErasureStatement struct {
	Version                    string    `json:"version"`
	TenantID                   string    `json:"tenantId"`
	ProjectID                  string    `json:"projectId"`
	EnvironmentID              string    `json:"environmentId"`
	ClusterID                  string    `json:"clusterId"`
	ClusterGeneration          int64     `json:"clusterGeneration"`
	NamespaceUID               string    `json:"namespaceUid"`
	PVCUID                     string    `json:"pvcUid"`
	PVUID                      string    `json:"pvUid"`
	BackendReference           string    `json:"backendReference"`
	Provisioner                string    `json:"provisioner"`
	Method                     string    `json:"method"`
	Coverage                   string    `json:"coverage"`
	CompletedAt                time.Time `json:"completedAt"`
	KeyExclusive               bool      `json:"keyExclusive"`
	KeyDestroyed               bool      `json:"keyDestroyed"`
	KeyCopiesDestroyed         bool      `json:"keyCopiesDestroyed"`
	EncryptedSinceProvisioning bool      `json:"encryptedSinceProvisioning"`
	PlaintextNeverWritten      bool      `json:"plaintextNeverWritten"`
	KeyBindingReference        string    `json:"keyBindingReference"`
}
