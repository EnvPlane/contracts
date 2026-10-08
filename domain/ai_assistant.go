package domain

// AIAssistantOutput is read-only narration. It never replaces a validated
// deterministic plan, action authorization, or an approval decision.
type AIAssistantOutput struct {
	Mode           string             `json:"mode"`
	Provider       string             `json:"provider,omitempty"`
	Model          string             `json:"model,omitempty"`
	RunID          string             `json:"runId,omitempty"`
	FallbackReason string             `json:"fallbackReason,omitempty"`
	Result         *AIDiagnosisResult `json:"result,omitempty"`
}

type AIAssistantSnapshot struct {
	TenantID  string
	ProjectID string
	SubjectID string
	Purpose   string
	Summary   string
}
