package domain

import "time"

type SAMLProviderConfig struct {
	TenantID         string            `json:"tenantId"`
	ProviderID       string            `json:"providerId"`
	Issuer           string            `json:"issuer"`
	Audience         string            `json:"audience"`
	MetadataRef      SecretReference   `json:"metadataRef"`
	SigningKeyRef    SecretReference   `json:"signingKeyRef"`
	SigningCertificatePEM string         `json:"signingCertificatePem"`
	Destination      string            `json:"destination"`
	SSOURL           string            `json:"ssoUrl"`
	AttributeMapping map[string]string `json:"attributeMapping"`
	Enabled          bool              `json:"enabled"`
}

type SAMLAssertion struct {
	AssertionID    string            `json:"assertionId"`
	Issuer         string            `json:"issuer"`
	Audience       string            `json:"audience"`
	Subject        string            `json:"subject"`
	Attributes     map[string]string `json:"attributes"`
	NotBefore      time.Time         `json:"notBefore"`
	NotOnOrAfter   time.Time         `json:"notOnOrAfter"`
}
