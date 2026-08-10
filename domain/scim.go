package domain

type SCIMUser struct {
	ID          string `json:"id"`
	ExternalID  string `json:"externalId,omitempty"`
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active"`
}

type SCIMGroup struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	Members     []string `json:"members,omitempty"`
	Role        string   `json:"role,omitempty"`
}

type SCIMListResponse[T any] struct {
	TotalResults int `json:"totalResults"`
	StartIndex   int `json:"startIndex"`
	ItemsPerPage int `json:"itemsPerPage"`
	Resources    []T `json:"Resources"`
}
