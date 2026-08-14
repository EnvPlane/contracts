package domain

import "time"

// AuthenticationSettings is the installation-wide, safe-to-read interactive
// login configuration. It intentionally has no clientSecret field.
type AuthenticationSettings struct {
	State              string                         `json:"state"`
	Mode               string                         `json:"mode"`
	Provider           *string                        `json:"provider,omitempty"`
	ProviderSettings   AuthenticationProviderSettings `json:"providerSettings"`
	Configured         bool                           `json:"configured"`
	CredentialRevision int64                          `json:"credentialRevision"`
	SessionRevision    int64                          `json:"sessionRevision"`
	UpdatedAt          *time.Time                     `json:"updatedAt,omitempty"`
	UpdatedBy          string                         `json:"updatedBy,omitempty"`
}

// AuthenticationProviderSettings contains only non-secret provider metadata.
type AuthenticationProviderSettings struct {
	ClientID         string   `json:"clientId,omitempty"`
	Issuer           string   `json:"issuer,omitempty"`
	AuthorizationURL string   `json:"authorizationUrl,omitempty"`
	TokenURL         string   `json:"tokenUrl,omitempty"`
	UserInfoURL      string   `json:"userInfoUrl,omitempty"`
	RevocationURL    string   `json:"revocationUrl,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	CallbackURL      string   `json:"callbackUrl,omitempty"`
}

// AuthenticationSettingsCommand is write-only transport input. clientSecret
// must never be copied into AuthenticationSettings, audit records, or logs.
type AuthenticationSettingsCommand struct {
	ExpectedCredentialRevision int64                          `json:"expectedCredentialRevision"`
	Mode                       string                         `json:"mode"`
	Provider                   *string                        `json:"provider,omitempty"`
	ProviderSettings           AuthenticationProviderSettings `json:"providerSettings,omitempty"`
	ClientSecret               string                         `json:"clientSecret,omitempty"`
}
