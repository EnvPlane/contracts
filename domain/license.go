package domain

import "time"

const LicenseSchemaVersion = "1"

type LicenseGrant struct {
	SchemaVersion string           `json:"schemaVersion"`
	LicenseID     string           `json:"licenseId"`
	CustomerID    string           `json:"customerId"`
	TenantID      string           `json:"tenantId"`
	PlanID        string           `json:"planId"`
	PlanVersion   string           `json:"planVersion"`
	Features      map[string]bool  `json:"features"`
	Limits        map[string]int64 `json:"limits"`
	IssuedAt      time.Time        `json:"issuedAt"`
	ExpiresAt     time.Time        `json:"expiresAt"`
}

type SignedLicense struct {
	KeyID     string       `json:"keyId"`
	Algorithm string       `json:"algorithm,omitempty"`
	Grant     LicenseGrant `json:"grant"`
	Signature string       `json:"signature"`
}
