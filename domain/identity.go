package domain

import "time"

// User identity is stable by provider plus provider subject; email is a
// mutable contact attribute and must never be used as a primary identity key.
type User struct {
	ID               string    `json:"id"`
	IdentityProvider string    `json:"identity_provider"`
	IdentitySubject  string    `json:"identity_subject"`
	Email            string    `json:"email,omitempty"`
	DisplayName      string    `json:"display_name,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	SessionEpoch     int64     `json:"session_epoch"`
}

type MembershipStatus string

const (
	MembershipActive   MembershipStatus = "active"
	MembershipDisabled MembershipStatus = "disabled"
)

type MembershipRole string

const (
	MembershipRoleOwner     MembershipRole = "owner"
	MembershipRoleAdmin     MembershipRole = "admin"
	MembershipRoleOperator  MembershipRole = "operator"
	MembershipRoleDeveloper MembershipRole = "developer"
	MembershipRoleViewer    MembershipRole = "viewer"
	MembershipRoleAuditor   MembershipRole = "auditor"
)

type Membership struct {
	TenantID  string           `json:"tenant_id"`
	UserID    string           `json:"user_id"`
	Status    MembershipStatus `json:"status"`
	Role      MembershipRole   `json:"role"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
