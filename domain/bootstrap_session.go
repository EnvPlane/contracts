package domain

import "time"

type BootstrapSessionStatus string

const (
	BootstrapSessionStatusDraft    BootstrapSessionStatus = "draft"
	BootstrapSessionStatusScanning BootstrapSessionStatus = "scanning"
	BootstrapSessionStatusReviewed BootstrapSessionStatus = "reviewed"
	BootstrapSessionStatusCompiled BootstrapSessionStatus = "compiled"
	BootstrapSessionStatusDeployed BootstrapSessionStatus = "deployed"
)

type BootstrapSession struct {
	TenantID    string                 `json:"tenant_id,omitempty"`
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	CurrentStep int                    `json:"current_step"`
	Status      BootstrapSessionStatus `json:"status"`
	CreatedBy   string                 `json:"created_by"`
	Data        map[string]any         `json:"data,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
