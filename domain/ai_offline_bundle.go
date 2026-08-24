package domain

import (
	"fmt"
	"strings"
)

const AIOfflineBundleSchemaVersion = "1"

// AIOfflineBundle is a signed-install artifact payload. Signature and artifact
// digest verification remain the responsibility of the air-gap installer.
type AIOfflineBundle struct {
	SchemaVersion string                              `json:"schemaVersion"`
	BundleID      string                              `json:"bundleId"`
	ModelID       string                              `json:"modelId"`
	Results       map[AIRequestKind]AIDiagnosisResult `json:"results"`
}

func (b AIOfflineBundle) Validate() error {
	if b.SchemaVersion != AIOfflineBundleSchemaVersion || strings.TrimSpace(b.BundleID) == "" || strings.TrimSpace(b.ModelID) == "" {
		return fmt.Errorf("invalid offline AI bundle identity or schema")
	}
	if len(b.Results) == 0 {
		return fmt.Errorf("offline AI bundle has no results")
	}
	for kind, result := range b.Results {
		if kind != AIRequestKindDiagnosis && kind != AIRequestKindBootstrap {
			return fmt.Errorf("unsupported offline AI request kind %q", kind)
		}
		result.RequestID = "bundle-template"
		result.TenantID = "bundle-template"
		if len(result.Evidence) != 0 {
			return fmt.Errorf("offline AI result %q must not contain fixed evidence", kind)
		}
		if err := result.Validate(); err != nil {
			return fmt.Errorf("offline AI result %q is invalid: %w", kind, err)
		}
	}
	return nil
}
