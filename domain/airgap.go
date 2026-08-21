package domain

import "time"

type AirgapArtifact struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	SourceRef    string `json:"sourceRef"`
	Digest       string `json:"digest"`
	Signature    string `json:"signature"`
	KeyID        string `json:"keyId,omitempty"`
	RegistryPath string `json:"registryPath"`
	ArchivePath  string `json:"archivePath,omitempty"`
}
type AirgapBundleManifest struct {
	BundleID          string            `json:"bundleId"`
	SchemaVersion     string            `json:"schemaVersion"`
	CreatedAt         time.Time         `json:"createdAt"`
	Artifacts         []AirgapArtifact  `json:"artifacts"`
	RegistryRemap     map[string]string `json:"registryRemap,omitempty"`
	LicenseRef        SecretReference   `json:"licenseRef"`
	ManifestKeyID     string            `json:"manifestKeyId,omitempty"`
	ManifestSignature string            `json:"manifestSignature,omitempty"`
	RunbookRef        string            `json:"runbookRef"`
}
