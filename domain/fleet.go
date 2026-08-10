package domain

type FleetGroup struct {
	ID                string            `json:"id"`
	TenantID          string            `json:"tenantId"`
	Name              string            `json:"name"`
	Selector          map[string]string `json:"selector"`
	MaintenanceWindow string            `json:"maintenanceWindow,omitempty"`
}
type UpgradeWave struct {
	ID                string `json:"id"`
	GroupID           string `json:"groupId"`
	MaxUnavailable    int    `json:"maxUnavailable"`
	ArtifactSet       string `json:"artifactSet"`
	PreviousSignedSet string `json:"previousSignedSet,omitempty"`
	Paused            bool   `json:"paused"`
	RolledBack        bool   `json:"rolledBack"`
}
type WavePlan struct {
	WaveID     string   `json:"waveId"`
	TenantID   string   `json:"tenantId"`
	ClusterIDs []string `json:"clusterIds"`
	Blocked    bool     `json:"blocked"`
	Reason     string   `json:"reason,omitempty"`
}
