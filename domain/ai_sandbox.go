package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const AISandboxSchemaVersion = "1"

type AISandboxTransform string

const (
	AISandboxSummarizeEvidence AISandboxTransform = "summarize_evidence"
	AISandboxCompareState      AISandboxTransform = "compare_structured_state"
	AISandboxExtractMetrics    AISandboxTransform = "extract_metrics"
)

type AISandboxResourceLimits struct {
	CPUMillis      int64 `json:"cpuMillis"`
	MemoryMiB      int64 `json:"memoryMiB"`
	PIDs           int   `json:"pids"`
	TimeoutSeconds int   `json:"timeoutSeconds"`
	MaxOutputBytes int   `json:"maxOutputBytes"`
}

type AISignedInputRef struct {
	ID        string `json:"id"`
	Digest    string `json:"digest"`
	Signature string `json:"signature"`
	KeyID     string `json:"keyId"`
}

type AISandboxEgress struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type AISandboxSpec struct {
	SchemaVersion string                  `json:"schemaVersion"`
	RunID         string                  `json:"runId"`
	TenantID      string                  `json:"tenantId"`
	ProjectID     string                  `json:"projectId"`
	Transform     AISandboxTransform      `json:"transform"`
	Image         string                  `json:"image"`
	Inputs        []AISignedInputRef      `json:"inputs"`
	Egress        []AISandboxEgress       `json:"egress"`
	NetworkDenied bool                    `json:"networkDenied"`
	Resources     AISandboxResourceLimits `json:"resources"`
	CreatedAt     time.Time               `json:"createdAt"`
}

type AISandboxOutput struct {
	SchemaVersion string `json:"schemaVersion"`
	RunID         string `json:"runId"`
	Digest        string `json:"digest"`
	Signature     string `json:"signature"`
	KeyID         string `json:"keyId"`
	Payload       string `json:"payload"`
}

var sandboxDigestPattern = regexp.MustCompile(`^.+@sha256:[a-f0-9]{64}$`)

func (r AISandboxResourceLimits) Validate() error {
	if r.CPUMillis < 1 || r.CPUMillis > 4000 || r.MemoryMiB < 16 || r.MemoryMiB > 8192 || r.PIDs < 1 || r.PIDs > 256 || r.TimeoutSeconds < 1 || r.TimeoutSeconds > 900 || r.MaxOutputBytes < 1 || r.MaxOutputBytes > 1024*1024 {
		return errors.New("AI sandbox resource limits are invalid or unbounded")
	}
	return nil
}

func (r AISignedInputRef) Validate() error {
	if strings.TrimSpace(r.ID) == "" || !strings.HasPrefix(r.Digest, "sha256:") || len(r.Digest) != 71 || strings.TrimSpace(r.Signature) == "" || strings.TrimSpace(r.KeyID) == "" {
		return errors.New("AI sandbox input reference must be signed and digest addressed")
	}
	return nil
}

func (s AISandboxSpec) Validate() error {
	if s.SchemaVersion != AISandboxSchemaVersion || strings.TrimSpace(s.RunID) == "" || strings.TrimSpace(s.TenantID) == "" || strings.TrimSpace(s.ProjectID) == "" || !sandboxDigestPattern.MatchString(s.Image) || !s.NetworkDenied || len(s.Inputs) == 0 || len(s.Inputs) > 32 || len(s.Egress) > 8 || s.CreatedAt.IsZero() {
		return errors.New("AI sandbox spec is invalid or unsafe")
	}
	switch s.Transform {
	case AISandboxSummarizeEvidence, AISandboxCompareState, AISandboxExtractMetrics:
	default:
		return errors.New("AI sandbox transform is not allowlisted")
	}
	if err := s.Resources.Validate(); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, input := range s.Inputs {
		if err := input.Validate(); err != nil {
			return err
		}
		if _, ok := seen[input.ID]; ok {
			return errors.New("AI sandbox input is duplicated")
		}
		seen[input.ID] = struct{}{}
	}
	for _, egress := range s.Egress {
		if strings.TrimSpace(egress.Host) == "" || egress.Port < 1 || egress.Port > 65535 || strings.ContainsAny(egress.Host, "/*?[]") {
			return errors.New("AI sandbox egress is not an exact destination")
		}
	}
	return nil
}

func (o AISandboxOutput) Validate(maxBytes int) error {
	if o.SchemaVersion != AISandboxSchemaVersion || strings.TrimSpace(o.RunID) == "" || strings.TrimSpace(o.Digest) == "" || strings.TrimSpace(o.Signature) == "" || strings.TrimSpace(o.KeyID) == "" || strings.TrimSpace(o.Payload) == "" || maxBytes <= 0 || len(o.Payload) > maxBytes {
		return errors.New("AI sandbox output is invalid or oversized")
	}
	return nil
}
