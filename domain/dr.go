package domain

import "time"

type BackupManifest struct {
	BackupID              string            `json:"backupId"`
	CreatedAt             time.Time         `json:"createdAt"`
	TenantIDs             []string          `json:"tenantIds"`
	PostgreSQLStateRef    string            `json:"postgresqlStateRef"`
	KubernetesMetadataRef string            `json:"kubernetesMetadataRef"`
	SecretReferences      []SecretReference `json:"secretReferences"`
	SigningKeyRef         SecretReference   `json:"signingKeyRef"`
	LicenseConfigRef      SecretReference   `json:"licenseConfigRef"`
	SchemaVersion         string            `json:"schemaVersion"`
}
type DRDrillReport struct {
	BackupID               string    `json:"backupId"`
	VerifiedAt             time.Time `json:"verifiedAt"`
	RestoreDurationSeconds int64     `json:"restoreDurationSeconds"`
	RPOSeconds             int64     `json:"rpoSeconds"`
	RTOSeconds             int64     `json:"rtoSeconds"`
	ConsistencyChecks      []string  `json:"consistencyChecks"`
	Passed                 bool      `json:"passed"`
	FailureReason          string    `json:"failureReason,omitempty"`
}
