package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const AIContextSchemaVersion = "1"

type AIContextTrust string

const AIContextTrustUntrustedData AIContextTrust = "untrusted_data"

type AIContextString struct {
	Value string         `json:"value"`
	Trust AIContextTrust `json:"trust"`
}

type AIContextField struct {
	Name  string          `json:"name"`
	Value AIContextString `json:"value"`
}

type AIContextEntry struct {
	SourceType string           `json:"sourceType"`
	SourceID   AIContextString  `json:"sourceId"`
	Fields     []AIContextField `json:"fields"`
}

type AIContext struct {
	SchemaVersion    string           `json:"schemaVersion"`
	TenantID         string           `json:"tenantId"`
	Entries          []AIContextEntry `json:"entries"`
	Truncated        bool             `json:"truncated"`
	TruncationMarker string           `json:"truncationMarker,omitempty"`
}

type AIContextLimits struct {
	MaxEntries     int
	MaxBytes       int
	MaxStringBytes int
}

func DefaultAIContextLimits() AIContextLimits {
	return AIContextLimits{MaxEntries: 256, MaxBytes: 64 * 1024, MaxStringBytes: 512}
}

type AIContextInput struct {
	TenantID     string
	Environments []Environment
	Jobs         []Job
	Events       []KubernetesEvent
	FluxStatuses []FluxStatus
	AgentHealth  []BootstrapAgentStatusResponse
	RunnerHealth []RunnerStatusResponse
	Capabilities []ClusterCapabilities
	Resources    []ResourceSnapshot
	Bootstrap    []AIBootstrapSnapshot
}

type AIContextThreatAssessment struct {
	CredentialDetected         bool `json:"credentialDetected"`
	PromptInjectionDetected    bool `json:"promptInjectionDetected"`
	ExfiltrationDetected       bool `json:"exfiltrationDetected"`
	UnicodeObfuscationDetected bool `json:"unicodeObfuscationDetected"`
}

type AISecurityEvent struct {
	SchemaVersion  string    `json:"schemaVersion"`
	EventType      string    `json:"eventType"`
	TenantID       string    `json:"tenantId"`
	SourceType     string    `json:"sourceType"`
	SourceIDHash   string    `json:"sourceIdHash"`
	Classification string    `json:"classification"`
	CreatedAt      time.Time `json:"createdAt"`
}

// AIBootstrapSnapshot is the explicit non-secret allowlist for troubleshooting.
type AIBootstrapSnapshot struct {
	TenantID              string `json:"tenantId"`
	ProjectID             string `json:"projectId"`
	SessionID             string `json:"sessionId"`
	CurrentStep           int    `json:"currentStep"`
	SessionStatus         string `json:"sessionStatus"`
	SCMStatus             string `json:"scmStatus"`
	AgentStatus           string `json:"agentStatus"`
	AgentError            string `json:"agentError,omitempty"`
	RunnerStatus          string `json:"runnerStatus"`
	RunnerError           string `json:"runnerError,omitempty"`
	ResourceScanStatus    string `json:"resourceScanStatus"`
	ResourceScanError     string `json:"resourceScanError,omitempty"`
	CapabilityReportStale bool   `json:"capabilityReportStale"`
	FailedCheck           string `json:"failedCheck,omitempty"`
	FailureMessage        string `json:"failureMessage,omitempty"`
}

type AIContextBuilder struct {
	limits AIContextLimits
}

func NewAIContextBuilder(limits AIContextLimits) AIContextBuilder {
	defaults := DefaultAIContextLimits()
	if limits.MaxEntries <= 0 {
		limits.MaxEntries = defaults.MaxEntries
	}
	if limits.MaxBytes <= 0 {
		limits.MaxBytes = defaults.MaxBytes
	}
	if limits.MaxStringBytes <= 0 {
		limits.MaxStringBytes = defaults.MaxStringBytes
	}
	return AIContextBuilder{limits: limits}
}

func (b AIContextBuilder) Build(input AIContextInput) (AIContext, error) {
	if strings.TrimSpace(input.TenantID) == "" {
		return AIContext{}, errors.New("AI context tenant is required")
	}
	context := AIContext{SchemaVersion: AIContextSchemaVersion, TenantID: input.TenantID}
	truncated := false
	appendEntry := func(entry AIContextEntry) {
		context.Entries = append(context.Entries, entry)
	}
	for _, environment := range input.Environments {
		if environment.TenantID != "" && environment.TenantID != input.TenantID {
			return AIContext{}, errors.New("AI environment tenant does not match context tenant")
		}
		appendEntry(environmentContextEntry(environment, b.limits.MaxStringBytes, &truncated))
	}
	for _, job := range input.Jobs {
		if job.TenantID != "" && job.TenantID != input.TenantID {
			return AIContext{}, errors.New("AI job tenant does not match context tenant")
		}
		appendEntry(jobContextEntry(job, b.limits.MaxStringBytes, &truncated))
	}
	for _, event := range input.Events {
		appendEntry(eventContextEntry(event, b.limits.MaxStringBytes, &truncated))
	}
	for _, flux := range input.FluxStatuses {
		appendEntry(fluxContextEntry(flux, b.limits.MaxStringBytes, &truncated))
	}
	for _, agent := range input.AgentHealth {
		appendEntry(agentContextEntry(agent, b.limits.MaxStringBytes, &truncated))
	}
	for _, runner := range input.RunnerHealth {
		appendEntry(runnerContextEntry(runner, b.limits.MaxStringBytes, &truncated))
	}
	for _, capabilities := range input.Capabilities {
		appendEntry(capabilityContextEntry(capabilities, b.limits.MaxStringBytes, &truncated))
	}
	for _, resource := range input.Resources {
		appendEntry(resourceContextEntry(resource, b.limits.MaxStringBytes, &truncated))
	}
	for _, snapshot := range input.Bootstrap {
		if snapshot.TenantID != input.TenantID {
			return AIContext{}, errors.New("AI bootstrap tenant does not match context tenant")
		}
		appendEntry(bootstrapContextEntry(snapshot, b.limits.MaxStringBytes, &truncated))
	}
	sort.Slice(context.Entries, func(i, j int) bool {
		left, right := context.Entries[i], context.Entries[j]
		if left.SourceType != right.SourceType {
			return left.SourceType < right.SourceType
		}
		return left.SourceID.Value < right.SourceID.Value
	})
	if len(context.Entries) > b.limits.MaxEntries {
		context.Entries = context.Entries[:b.limits.MaxEntries]
		truncated = true
	}
	context.Truncated = truncated
	if truncated {
		context.TruncationMarker = "AI_CONTEXT_TRUNCATED_MAX_ENTRIES_OR_STRING"
	}
	for len(context.Entries) > 0 {
		encoded, err := json.Marshal(context)
		if err != nil {
			return AIContext{}, fmt.Errorf("marshal AI context: %w", err)
		}
		if len(encoded) <= b.limits.MaxBytes {
			return context, nil
		}
		context.Entries = context.Entries[:len(context.Entries)-1]
		context.Truncated = true
		context.TruncationMarker = "AI_CONTEXT_TRUNCATED_MAX_BYTES"
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		return AIContext{}, fmt.Errorf("marshal AI context: %w", err)
	}
	if len(encoded) > b.limits.MaxBytes {
		return AIContext{}, errors.New("AI context byte limit is too small for its envelope")
	}
	return context, nil
}

func (c AIContext) SecureForProvider() (AIContext, AIContextThreatAssessment, error) {
	if c.SchemaVersion != AIContextSchemaVersion || strings.TrimSpace(c.TenantID) == "" {
		return AIContext{}, AIContextThreatAssessment{}, errors.New("AI context security envelope is invalid")
	}
	secured := c
	secured.Entries = make([]AIContextEntry, len(c.Entries))
	for index, entry := range c.Entries {
		secured.Entries[index] = entry
		secured.Entries[index].Fields = append([]AIContextField(nil), entry.Fields...)
	}
	assessment := AIContextThreatAssessment{}
	for entryIndex := range secured.Entries {
		secured.Entries[entryIndex].SourceID, assessment = secureContextString(secured.Entries[entryIndex].SourceID, assessment)
		for fieldIndex := range secured.Entries[entryIndex].Fields {
			secured.Entries[entryIndex].Fields[fieldIndex].Value, assessment = secureContextString(secured.Entries[entryIndex].Fields[fieldIndex].Value, assessment)
		}
	}
	return secured, assessment, nil
}

func bootstrapContextEntry(snapshot AIBootstrapSnapshot, maxBytes int, truncated *bool) AIContextEntry {
	return aiEntry("bootstrap_session", snapshot.SessionID, []AIContextField{
		aiField("projectId", snapshot.ProjectID, maxBytes, truncated),
		aiField("currentStep", strconv.Itoa(snapshot.CurrentStep), maxBytes, truncated),
		aiField("sessionStatus", snapshot.SessionStatus, maxBytes, truncated),
		aiField("scmStatus", snapshot.SCMStatus, maxBytes, truncated),
		aiField("agentStatus", snapshot.AgentStatus, maxBytes, truncated),
		aiField("agentError", snapshot.AgentError, maxBytes, truncated),
		aiField("runnerStatus", snapshot.RunnerStatus, maxBytes, truncated),
		aiField("runnerError", snapshot.RunnerError, maxBytes, truncated),
		aiField("resourceScanStatus", snapshot.ResourceScanStatus, maxBytes, truncated),
		aiField("resourceScanError", snapshot.ResourceScanError, maxBytes, truncated),
		aiField("capabilityReportStale", strconv.FormatBool(snapshot.CapabilityReportStale), maxBytes, truncated),
		aiField("failedCheck", snapshot.FailedCheck, maxBytes, truncated),
		aiField("failureMessage", snapshot.FailureMessage, maxBytes, truncated),
	})
}

var (
	aiBearerPattern     = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+`)
	aiPrivateKeyPattern = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	aiCredentialPattern = regexp.MustCompile(`(?i)\b(password|passwd|client[_-]?secret|webhook[_-]?secret|access[_-]?token|refresh[_-]?token|token)\s*[:=]\s*[^\s,;]+`)
)

func redactAIText(value string, maxBytes int) (AIContextString, bool) {
	value = strings.Map(func(r rune) rune {
		if r == '\u200b' || r == '\u200c' || r == '\u200d' || r == '\ufeff' {
			return -1
		}
		return r
	}, value)
	value = aiBearerPattern.ReplaceAllString(value, "Bearer [REDACTED]")
	value = aiPrivateKeyPattern.ReplaceAllString(value, "[REDACTED_PRIVATE_KEY]")
	value = aiCredentialPattern.ReplaceAllString(value, "$1=[REDACTED]")
	if strings.Contains(value, "apiVersion: v1") && (strings.Contains(value, "kind: Config") || strings.Contains(value, "current-context:")) {
		value = "[REDACTED_KUBECONFIG]"
	}
	if len(value) <= maxBytes {
		return AIContextString{Value: value, Trust: AIContextTrustUntrustedData}, false
	}
	marker := "[TRUNCATED_STRING]"
	if maxBytes <= len(marker) {
		return AIContextString{Value: marker[:maxBytes], Trust: AIContextTrustUntrustedData}, true
	}
	return AIContextString{Value: value[:maxBytes-len(marker)] + marker, Trust: AIContextTrustUntrustedData}, true
}

func secureContextString(value AIContextString, assessment AIContextThreatAssessment) (AIContextString, AIContextThreatAssessment) {
	normalized := strings.Map(func(r rune) rune {
		if r == '\u200b' || r == '\u200c' || r == '\u200d' || r == '\ufeff' {
			return -1
		}
		return r
	}, value.Value)
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, "bearer ") || strings.Contains(lower, "private key") || strings.Contains(lower, "password=") || strings.Contains(lower, "token=") || strings.Contains(lower, "webhook_secret") {
		assessment.CredentialDetected = true
	}
	if strings.Contains(lower, "ignore previous") || strings.Contains(lower, "system message") || strings.Contains(lower, "you are now") || strings.Contains(lower, "follow these instructions") {
		assessment.PromptInjectionDetected = true
	}
	if strings.Contains(lower, "send the secret") || strings.Contains(lower, "exfiltrate") || strings.Contains(lower, "upload credentials") {
		assessment.ExfiltrationDetected = true
	}
	if normalized != value.Value {
		assessment.UnicodeObfuscationDetected = true
	}
	redacted, _ := redactAIText(normalized, len(normalized))
	redacted.Trust = AIContextTrustUntrustedData
	return redacted, assessment
}

func SanitizeAIDiagnosisResult(result AIDiagnosisResult) (AIDiagnosisResult, AIContextThreatAssessment) {
	assessment := AIContextThreatAssessment{}
	result.Summary, assessment = sanitizeOutputText(result.Summary, assessment)
	for index := range result.Observed {
		result.Observed[index], assessment = sanitizeOutputText(result.Observed[index], assessment)
	}
	for index := range result.LikelyCauses {
		result.LikelyCauses[index].Summary, assessment = sanitizeOutputText(result.LikelyCauses[index].Summary, assessment)
	}
	for index := range result.SafeChecks {
		result.SafeChecks[index].Summary, assessment = sanitizeOutputText(result.SafeChecks[index].Summary, assessment)
	}
	for index := range result.UserChecks {
		result.UserChecks[index].Summary, assessment = sanitizeOutputText(result.UserChecks[index].Summary, assessment)
	}
	for index := range result.PlatformChecks {
		result.PlatformChecks[index].Summary, assessment = sanitizeOutputText(result.PlatformChecks[index].Summary, assessment)
	}
	return result, assessment
}

func sanitizeOutputText(value string, assessment AIContextThreatAssessment) (string, AIContextThreatAssessment) {
	sanitized, next := secureContextString(AIContextString{Value: value, Trust: AIContextTrustUntrustedData}, assessment)
	return sanitized.Value, next
}

// RedactAIText is the shared redaction boundary for all AI-derived storage.
// Callers must treat the returned value as untrusted data.
func RedactAIText(value string, maxBytes int) (string, bool) {
	redacted, truncated := redactAIText(value, maxBytes)
	return redacted.Value, truncated
}

func aiField(name, value string, maxBytes int, truncated *bool) AIContextField {
	redacted, wasTruncated := redactAIText(value, maxBytes)
	if wasTruncated {
		*truncated = true
	}
	return AIContextField{Name: name, Value: redacted}
}

func aiEntry(sourceType, sourceID string, fields []AIContextField) AIContextEntry {
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].Name != fields[j].Name {
			return fields[i].Name < fields[j].Name
		}
		return fields[i].Value.Value < fields[j].Value.Value
	})
	return AIContextEntry{SourceType: sourceType, SourceID: AIContextString{Value: sourceID, Trust: AIContextTrustUntrustedData}, Fields: fields}
}

func environmentContextEntry(environment Environment, maxBytes int, truncated *bool) AIContextEntry {
	return aiEntry("environment", environment.ID, []AIContextField{
		aiField("project", environment.Project, maxBytes, truncated),
		aiField("status", string(environment.Status), maxBytes, truncated),
		aiField("mode", string(environment.Mode), maxBytes, truncated),
		aiField("namespace", environment.Namespace, maxBytes, truncated),
		aiField("clusterId", environment.ClusterID, maxBytes, truncated),
		aiField("sourceRepository", environment.Source.Repository, maxBytes, truncated),
		aiField("sourceBranch", environment.Source.Branch, maxBytes, truncated),
		aiField("sourceProvider", environment.Source.Provider, maxBytes, truncated),
		aiField("ttlHours", strconv.Itoa(environment.TTLHours), maxBytes, truncated),
		aiField("pinned", strconv.FormatBool(environment.Pinned), maxBytes, truncated),
	})
}

func jobContextEntry(job Job, maxBytes int, truncated *bool) AIContextEntry {
	return aiEntry("job", job.ID, []AIContextField{
		aiField("type", string(job.Type), maxBytes, truncated),
		aiField("status", string(job.Status), maxBytes, truncated),
		aiField("environmentId", job.EnvironmentID, maxBytes, truncated),
		aiField("attempts", strconv.Itoa(job.Attempts), maxBytes, truncated),
		aiField("maxAttempts", strconv.Itoa(job.MaxAttempts), maxBytes, truncated),
	})
}

func eventContextEntry(event KubernetesEvent, maxBytes int, truncated *bool) AIContextEntry {
	sourceID := event.UID
	if sourceID == "" {
		sourceID = event.Namespace + "/" + event.InvolvedKind + "/" + event.InvolvedName + "/" + event.LastSeen.UTC().Format(time.RFC3339Nano)
	}
	return aiEntry("kubernetes_event", sourceID, []AIContextField{
		aiField("namespace", event.Namespace, maxBytes, truncated),
		aiField("type", event.Type, maxBytes, truncated),
		aiField("reason", event.Reason, maxBytes, truncated),
		aiField("message", event.Message, maxBytes, truncated),
		aiField("involvedKind", event.InvolvedKind, maxBytes, truncated),
		aiField("involvedName", event.InvolvedName, maxBytes, truncated),
		aiField("count", strconv.FormatInt(int64(event.Count), 10), maxBytes, truncated),
	})
}

func fluxContextEntry(flux FluxStatus, maxBytes int, truncated *bool) AIContextEntry {
	fields := []AIContextField{aiField("status", string(flux.Status), maxBytes, truncated), aiField("message", flux.Message, maxBytes, truncated)}
	for _, resource := range append(append([]FluxResourceStatus{}, flux.Kustomizations...), flux.HelmReleases...) {
		fields = append(fields,
			aiField("resource.kind", resource.Kind, maxBytes, truncated),
			aiField("resource.name", resource.Name, maxBytes, truncated),
			aiField("resource.namespace", resource.Namespace, maxBytes, truncated),
			aiField("resource.ready", strconv.FormatBool(resource.Ready), maxBytes, truncated),
			aiField("resource.failed", strconv.FormatBool(resource.Failed), maxBytes, truncated),
			aiField("resource.reason", resource.Reason, maxBytes, truncated),
		)
	}
	return aiEntry("flux_status", flux.UpdatedAt.UTC().Format(time.RFC3339Nano), fields)
}

func agentContextEntry(agent BootstrapAgentStatusResponse, maxBytes int, truncated *bool) AIContextEntry {
	return aiEntry("agent_health", agent.AgentID, []AIContextField{
		aiField("status", agent.Status, maxBytes, truncated),
		aiField("effectiveStatus", agent.EffectiveStatus, maxBytes, truncated),
		aiField("statusReason", agent.StatusReason, maxBytes, truncated),
		aiField("clusterId", agent.ClusterID, maxBytes, truncated),
		aiField("targetClusterMode", agent.TargetClusterMode, maxBytes, truncated),
		aiField("resourceScanStatus", agent.ResourceScanStatus, maxBytes, truncated),
		aiField("resourceCount", strconv.Itoa(agent.ResourceCount), maxBytes, truncated),
	})
}

func runnerContextEntry(runner RunnerStatusResponse, maxBytes int, truncated *bool) AIContextEntry {
	return aiEntry("runner_health", runner.RunnerID, []AIContextField{
		aiField("status", runner.Status, maxBytes, truncated),
		aiField("effectiveStatus", runner.EffectiveStatus, maxBytes, truncated),
		aiField("statusReason", runner.StatusReason, maxBytes, truncated),
		aiField("deploymentMode", runner.DeploymentMode, maxBytes, truncated),
		aiField("clusterId", runner.ClusterID, maxBytes, truncated),
		aiField("runnerNamespace", runner.RunnerNamespace, maxBytes, truncated),
		aiField("targetClusterMode", runner.TargetClusterMode, maxBytes, truncated),
		aiField("tokenState", runner.TokenState, maxBytes, truncated),
		aiField("recoveryAction", runner.RecoveryAction, maxBytes, truncated),
	})
}

func capabilityContextEntry(capabilities ClusterCapabilities, maxBytes int, truncated *bool) AIContextEntry {
	fields := []AIContextField{aiField("kubernetesVersion", capabilities.KubernetesVersion, maxBytes, truncated)}
	for _, value := range append(append([]string{}, capabilities.Capabilities...), capabilities.Report.CapabilityFlags...) {
		fields = append(fields, aiField("capability", value, maxBytes, truncated))
	}
	fields = append(fields,
		aiField("revision", strconv.Itoa(capabilities.Report.Revision), maxBytes, truncated),
		aiField("namespaceMode", capabilities.Report.NamespaceMode, maxBytes, truncated),
		aiField("externalDNSPresent", strconv.FormatBool(capabilities.Report.ExternalDNSPresent), maxBytes, truncated),
	)
	return aiEntry("capability_report", capabilities.Report.ConfigFingerprint, fields)
}

func resourceContextEntry(resource ResourceSnapshot, maxBytes int, truncated *bool) AIContextEntry {
	labelKeys := make([]string, 0, len(resource.Labels))
	for key := range resource.Labels {
		labelKeys = append(labelKeys, key)
	}
	sort.Strings(labelKeys)
	fields := []AIContextField{
		aiField("kind", resource.Kind, maxBytes, truncated),
		aiField("namespace", resource.Namespace, maxBytes, truncated),
		aiField("labelKeys", strings.Join(labelKeys, ","), maxBytes, truncated),
	}
	if resource.Health != nil {
		fields = append(fields, aiField("health.status", resource.Health.Status, maxBytes, truncated), aiField("health.message", resource.Health.Message, maxBytes, truncated))
	}
	return aiEntry("resource_metadata", resource.Kind+"/"+resource.Namespace+"/"+resource.Name, fields)
}
