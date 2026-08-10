package domain

import "time"

type ApprovalState string

const (
	ApprovalPending  ApprovalState = "pending"
	ApprovalApproved ApprovalState = "approved"
	ApprovalRejected ApprovalState = "rejected"
	ApprovalExpired  ApprovalState = "expired"
	ApprovalCanceled ApprovalState = "canceled"
)

type ApprovalRequest struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenantId"`
	RequesterID string        `json:"requesterId"`
	Action      string        `json:"action"`
	ResourceID  string        `json:"resourceId"`
	State       ApprovalState `json:"state"`
	ExpiresAt   time.Time     `json:"expiresAt"`
	ApprovedBy  string        `json:"approvedBy,omitempty"`
	DecisionAt  *time.Time    `json:"decisionAt,omitempty"`
	ResumeKey   string        `json:"resumeKey"`
}
