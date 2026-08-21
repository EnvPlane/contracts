package domain

type FleetGroup struct {
	ID                 string                   `json:"id"`
	TenantID           string                   `json:"tenantId"`
	Name               string                   `json:"name"`
	Labels             map[string]string        `json:"labels,omitempty"`
	Selector           map[string]string        `json:"selector"`
	MaintenanceWindow  string                   `json:"maintenanceWindow,omitempty"`
	MaintenanceWindows []FleetMaintenanceWindow `json:"maintenanceWindows,omitempty"`
}

type FleetMaintenanceWindow struct {
	Weekday         int `json:"weekday"`
	StartMinuteUTC  int `json:"startMinuteUtc"`
	DurationMinutes int `json:"durationMinutes"`
}

type SignedArtifactSet struct {
	Reference string `json:"reference"`
	Digest    string `json:"digest"`
	Signature string `json:"signature"`
	KeyID     string `json:"keyId"`
}

type UpgradeWave struct {
	ID                string             `json:"id"`
	GroupID           string             `json:"groupId"`
	MaxUnavailable    int                `json:"maxUnavailable"`
	ArtifactSet       string             `json:"artifactSet"`
	PreviousSignedSet string             `json:"previousSignedSet,omitempty"`
	SignedArtifact    *SignedArtifactSet `json:"signedArtifact,omitempty"`
	AgentVersion      string             `json:"agentVersion,omitempty"`
	RunnerVersion     string             `json:"runnerVersion,omitempty"`
	Status            string             `json:"status,omitempty"`
	Paused            bool               `json:"paused"`
	RolledBack        bool               `json:"rolledBack"`
}
type WavePlan struct {
	WaveID                string   `json:"waveId"`
	TenantID              string   `json:"tenantId"`
	ClusterIDs            []string `json:"clusterIds"`
	ArtifactSet           string   `json:"artifactSet,omitempty"`
	AgentVersion          string   `json:"agentVersion,omitempty"`
	RunnerVersion         string   `json:"runnerVersion,omitempty"`
	MaintenanceWindowOpen bool     `json:"maintenanceWindowOpen"`
	HealthGatePassed      bool     `json:"healthGatePassed"`
	Blocked               bool     `json:"blocked"`
	Reason                string   `json:"reason,omitempty"`
}
