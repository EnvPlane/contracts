package domain

import "time"

type BackupManifest struct {
	BackupID              string            `json:"backupId"`
	CreatedAt             time.Time         `json:"createdAt"`
	TenantIDs             []string          `json:"tenantIds"`
	ProjectIDs            []string          `json:"projectIds,omitempty"`
	PostgreSQLStateRef    string            `json:"postgresqlStateRef"`
	KubernetesMetadataRef string            `json:"kubernetesMetadataRef"`
	SubscriptionStateRef  string            `json:"subscriptionStateRef,omitempty"`
	AuditStateRef         string            `json:"auditStateRef,omitempty"`
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
	SecretRebindRequired   bool      `json:"secretRebindRequired"`
	Passed                 bool      `json:"passed"`
	FailureReason          string    `json:"failureReason,omitempty"`
}

type DRDrillSchedule struct {
	TenantID       string    `json:"tenantId"`
	IntervalHours  int       `json:"intervalHours"`
	LastVerifiedAt time.Time `json:"lastVerifiedAt,omitempty"`
	NextDueAt      time.Time `json:"nextDueAt,omitempty"`
}

func (s DRDrillSchedule) Due(now time.Time) bool {
	if s.IntervalHours <= 0 || s.LastVerifiedAt.IsZero() {
		return true
	}
	return !now.UTC().Before(s.LastVerifiedAt.UTC().Add(time.Duration(s.IntervalHours) * time.Hour))
}

func (s DRDrillSchedule) WithVerifiedAt(at time.Time) DRDrillSchedule {
	s.LastVerifiedAt = at.UTC()
	if s.IntervalHours > 0 {
		s.NextDueAt = s.LastVerifiedAt.Add(time.Duration(s.IntervalHours) * time.Hour)
	}
	return s
}
