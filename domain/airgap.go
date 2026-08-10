package domain

import "time"

type AirgapArtifact struct {
	Name         string `json:"name"`
	SourceRef    string `json:"sourceRef"`
	Digest       string `json:"digest"`
	Signature    string `json:"signature"`
	RegistryPath string `json:"registryPath"`
}
type AirgapBundleManifest struct {
	BundleID      string           `json:"bundleId"`
	SchemaVersion string           `json:"schemaVersion"`
	CreatedAt     time.Time        `json:"createdAt"`
	Artifacts     []AirgapArtifact `json:"artifacts"`
	LicenseRef    SecretReference  `json:"licenseRef"`
	RunbookRef    string           `json:"runbookRef"`
}
