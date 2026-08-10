package domain

type OIDCProviderConfig struct {
	TenantID         string          `json:"tenantId"`
	ProviderID       string          `json:"providerId"`
	Issuer           string          `json:"issuer"`
	ClientID         string          `json:"clientId"`
	ClientSecretRef  SecretReference `json:"clientSecretRef"`
	AuthorizationURL string          `json:"authorizationUrl"`
	TokenURL         string          `json:"tokenUrl"`
	UserInfoURL      string          `json:"userInfoUrl"`
	Scopes           []string        `json:"scopes"`
	Audience         string          `json:"audience,omitempty"`
	Enabled          bool            `json:"enabled"`
}
